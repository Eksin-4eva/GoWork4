package service

import (
	"net/http"
	"strconv"
	"time"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
)

// historyItem 是一条历史消息。
type historyItem struct {
	ID         string `json:"id"`
	RoomID     string `json:"room_id"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

// HandleHistory 提供与某个用户的会话历史。
//
//	GET /history?peer_id=<对方用户 ID>&page_num=1&page_size=20
//
// 旧版只有 WebSocket，没有拉历史的能力；一旦断线重连，之前的消息就再也看不到了。
func (s *Service) HandleHistory(w http.ResponseWriter, r *http.Request) {
	userID, tokens, err := s.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, errno.AuthInvalid)
		return
	}
	for key, value := range tokens {
		w.Header().Set(key, value)
	}

	peerID, err := strconv.ParseInt(r.URL.Query().Get("peer_id"), 10, 64)
	if err != nil || peerID <= 0 {
		writeError(w, http.StatusOK, errno.ParamError.WithMessage("peer_id 不合法"))
		return
	}

	pageNum, _ := strconv.Atoi(r.URL.Query().Get("page_num"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = constants.ChatHistoryPageSize
	}

	messages, total, err := s.db.Chat.ListByRoom(
		r.Context(), buildRoomID(userID, peerID), pageNum, pageSize)
	if err != nil {
		writeError(w, http.StatusOK, err)
		return
	}

	items := make([]historyItem, 0, len(messages))
	for _, m := range messages {
		items = append(items, historyItem{
			ID:         strconv.FormatInt(m.ID, 10),
			RoomID:     m.RoomID,
			FromUserID: strconv.FormatInt(m.SenderID, 10),
			ToUserID:   strconv.FormatInt(m.ReceiverID, 10),
			Content:    m.Content,
			CreatedAt:  m.CreatedAt.Format(time.DateTime),
		})
	}

	writeData(w, map[string]any{"items": items, "total": total})
}
