package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gobili/chat/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Service struct {
	cfg      config.Config
	db       *sql.DB
	upgrader websocket.Upgrader

	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
}

type Client struct {
	userID string
	conn   *websocket.Conn
	mu     sync.Mutex
}

type incomingMessage struct {
	ToUserID    string `json:"to_user_id"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

type outgoingMessage struct {
	Type      string `json:"type"`
	ID        string `json:"id,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
	FromUser  string `json:"from_user_id,omitempty"`
	ToUser    string `json:"to_user_id,omitempty"`
	Content   string `json:"content,omitempty"`
	SentAt    string `json:"sent_at,omitempty"`
	Error     string `json:"error,omitempty"`
	Online    bool   `json:"online,omitempty"`
	Refreshed bool   `json:"refreshed,omitempty"`
}

func New(cfg config.Config) (*Service, error) {
	db, err := sql.Open("mysql", cfg.MySQL.DSN)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Service{
		cfg: cfg,
		db:  db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
		clients: make(map[string]map[*Client]struct{}),
	}, nil
}

func (s *Service) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Service) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, headers, err := s.authenticate(r.Context(), r.Header.Get(s.cfg.Auth.AccessHeader), r.Header.Get(s.cfg.Auth.RefreshHeader))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	responseHeader := http.Header{}
	for key, value := range headers {
		if value != "" {
			responseHeader.Set(key, value)
		}
	}
	conn, err := s.upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return
	}

	client := &Client{userID: userID, conn: conn}
	s.addClient(client)
	defer s.removeClient(client)
	defer conn.Close()

	_ = client.writeJSON(outgoingMessage{Type: "connected", Refreshed: headers[s.cfg.Auth.AccessHeader] != r.Header.Get(s.cfg.Auth.AccessHeader)})

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req incomingMessage
		if err := json.Unmarshal(payload, &req); err != nil {
			_ = client.writeJSON(outgoingMessage{Type: "error", Error: "invalid message payload"})
			continue
		}
		if err := s.handleMessage(r.Context(), client, req); err != nil {
			_ = client.writeJSON(outgoingMessage{Type: "error", Error: err.Error()})
		}
	}
}

func (s *Service) handleMessage(ctx context.Context, client *Client, req incomingMessage) error {
	toUserID := strings.TrimSpace(req.ToUserID)
	if toUserID == "" {
		return errors.New("invalid to_user_id")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return errors.New("content is required")
	}
	messageType := strings.TrimSpace(req.MessageType)
	if messageType == "" {
		messageType = "text"
	}

	messageID := fmt.Sprintf("%d", time.Now().UnixNano())
	roomID := buildRoomID(client.userID, toUserID)
	now := time.Now()
	if err := s.storeMessage(ctx, messageID, roomID, client.userID, toUserID, messageType, content); err != nil {
		return err
	}

	message := outgoingMessage{
		Type:     "message",
		ID:       messageID,
		RoomID:   roomID,
		FromUser: client.userID,
		ToUser:   toUserID,
		Content:  content,
		SentAt:   now.Format("2006-01-02 15:04:05"),
	}
	online := s.broadcastToUser(toUserID, message)
	_ = client.writeJSON(outgoingMessage{
		Type:     "ack",
		ID:       message.ID,
		RoomID:   message.RoomID,
		FromUser: message.FromUser,
		ToUser:   message.ToUser,
		Content:  message.Content,
		SentAt:   message.SentAt,
		Online:   online,
	})
	return nil
}

func (s *Service) authenticate(ctx context.Context, accessToken, refreshToken string) (string, map[string]string, error) {
	if accessToken == "" && refreshToken == "" {
		return "", nil, errors.New("missing token")
	}
	if accessToken != "" {
		userID, err := s.parseToken(accessToken, s.cfg.JWT.AccessSecret)
		if err == nil {
			return userID, map[string]string{
				s.cfg.Auth.AccessHeader:  accessToken,
				s.cfg.Auth.RefreshHeader: refreshToken,
			}, nil
		}
	}
	if refreshToken == "" {
		return "", nil, errors.New("invalid token")
	}
	userID, err := s.parseToken(refreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return "", nil, err
	}
	valid, err := s.isRefreshTokenValid(ctx, refreshToken)
	if err != nil {
		return "", nil, err
	}
	if !valid {
		return "", nil, errors.New("refresh token revoked")
	}
	newAccess, err := s.signToken(userID, s.cfg.JWT.AccessSecret, s.cfg.JWT.AccessExpireSeconds)
	if err != nil {
		return "", nil, err
	}
	return userID, map[string]string{
		s.cfg.Auth.AccessHeader:  newAccess,
		s.cfg.Auth.RefreshHeader: refreshToken,
	}, nil
}

func (s *Service) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clients[client.userID] == nil {
		s.clients[client.userID] = make(map[*Client]struct{})
	}
	s.clients[client.userID][client] = struct{}{}
}

func (s *Service) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	group := s.clients[client.userID]
	delete(group, client)
	if len(group) == 0 {
		delete(s.clients, client.userID)
	}
}

func (s *Service) broadcastToUser(userID string, payload outgoingMessage) bool {
	s.mu.RLock()
	clients := s.clients[userID]
	copies := make([]*Client, 0, len(clients))
	for client := range clients {
		copies = append(copies, client)
	}
	s.mu.RUnlock()
	if len(copies) == 0 {
		return false
	}
	for _, client := range copies {
		if err := client.writeJSON(payload); err != nil {
			client.conn.Close()
			s.removeClient(client)
		}
	}
	return true
}

func (s *Service) storeMessage(ctx context.Context, id string, roomID string, senderID, receiverID string, messageType, content string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO chat_messages (id, room_id, sender_id, receiver_id, message_type, content) VALUES (?, ?, ?, ?, ?, ?)`, id, roomID, senderID, receiverID, messageType, content)
	return err
}

func (s *Service) isRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM refresh_tokens WHERE token = ? AND revoked = 0 AND deleted_at IS NULL AND expires_at > NOW()`, token).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) signToken(userID string, secret string, expireSeconds int64) (string, error) {
	claims := jwt.MapClaims{"user_id": userID, "exp": time.Now().Add(time.Duration(expireSeconds) * time.Second).Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (s *Service) parseToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	value, ok := claims["user_id"]
	if !ok {
		return "", errors.New("missing user_id")
	}
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprint(v), nil
	}
}

func (c *Client) writeJSON(v outgoingMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(v)
}

func buildRoomID(a, b string) string {
	if a < b {
		return a + "_" + b
	}
	return b + "_" + a
}
