package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// incomingMessage 是客户端发来的消息。
type incomingMessage struct {
	ToUserID    string `json:"to_user_id"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

// outgoingMessage 是服务端下发的消息。type 取值：
//
//	connected 握手成功
//	message   收到新消息
//	ack       自己发出的消息已落库
//	error     处理失败
type outgoingMessage struct {
	Type     string `json:"type"`
	ID       string `json:"id,omitempty"`
	RoomID   string `json:"room_id,omitempty"`
	FromUser string `json:"from_user_id,omitempty"`
	ToUser   string `json:"to_user_id,omitempty"`
	Content  string `json:"content,omitempty"`
	SentAt   string `json:"sent_at,omitempty"`
	Online   bool   `json:"online,omitempty"`
	Error    string `json:"error,omitempty"`
}

// buildRoomID 生成会话 ID：双方 ID 排序后拼接，
// 保证 A→B 与 B→A 落在同一个会话里。
func buildRoomID(a, b int64) string {
	if a > b {
		a, b = b, a
	}
	return strconv.FormatInt(a, 10) + "_" + strconv.FormatInt(b, 10)
}

// handleMessage 处理一条客户端消息：校验 → 落库 → 投递 → 回执。
func (s *Service) handleMessage(
	ctx context.Context, from int64, sess *session, req incomingMessage,
) error {
	toUserID, err := strconv.ParseInt(strings.TrimSpace(req.ToUserID), 10, 64)
	if err != nil || toUserID <= 0 {
		return errno.ParamError.WithMessage("to_user_id 不合法")
	}
	if toUserID == from {
		return errno.ParamError.WithMessage("不能给自己发消息")
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return errno.ParamEmpty.WithMessage("消息内容不能为空")
	}
	if len([]rune(content)) > constants.MaxMessageRunes {
		return errno.ParamError.WithMessage("消息过长")
	}

	// 收件人必须存在。旧版完全没校验，任意 ID 都能发，
	// 于是会话表里会堆满指向不存在用户的消息。
	found, _, err := s.db.User.GetByID(ctx, toUserID)
	if err != nil {
		return err
	}
	if !found {
		return errno.BizNotExist.WithMessage("接收方不存在")
	}

	messageType := strings.TrimSpace(req.MessageType)
	if messageType == "" {
		messageType = constants.DefaultMessageType
	}

	id, err := s.sf.NextVal()
	if err != nil {
		logger.Errorf("chat.handleMessage: snowflake: %v", err)
		return errno.InternalError
	}

	roomID := buildRoomID(from, toUserID)
	now := time.Now()
	if err := s.db.Chat.Save(ctx, &model.ChatMessage{
		ID:          id,
		RoomID:      roomID,
		SenderID:    from,
		ReceiverID:  toUserID,
		MessageType: messageType,
		Content:     content,
	}); err != nil {
		return err
	}

	message := outgoingMessage{
		Type:     "message",
		ID:       strconv.FormatInt(id, 10),
		RoomID:   roomID,
		FromUser: strconv.FormatInt(from, 10),
		ToUser:   strconv.FormatInt(toUserID, 10),
		Content:  content,
		SentAt:   now.Format("2006-01-02 15:04:05"),
	}
	online := s.hub.deliver(toUserID, message)

	// 给发送方回执：对方不在线也没关系，消息已落库，之后拉历史能看到
	ack := message
	ack.Type = "ack"
	ack.Online = online
	return sess.writeJSON(ack)
}
