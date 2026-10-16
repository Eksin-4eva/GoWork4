package service

import (
	"net/http"
	"strconv"
	"time"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
)

// conversationItem 是会话列表中的一条摘要。
type conversationItem struct {
	PeerID         string `json:"peer_id"`
	RoomID         string `json:"room_id"`
	LastMessageID  string `json:"last_message_id"`
	LastContent    string `json:"last_content"`
	LastFromUserID string `json:"last_from_user_id"`
	LastAt         string `json:"last_at"`
}

// HandleConversations 返回当前用户的会话列表。
//
//	GET /conversations?page_num=1&page_size=20
//
// 每条记录对应该会话的最新一条消息，按时间倒序。
func (s *Service) HandleConversations(w http.ResponseWriter, r *http.Request) {
	userID, tokens, err := s.authenticate(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, errno.AuthInvalid)
		return
	}
	for key, value := range tokens {
		w.Header().Set(key, value)
	}

	pageNum, _ := strconv.Atoi(r.URL.Query().Get("page_num"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = constants.ChatHistoryPageSize
	}

	messages, total, err := s.db.Chat.ListConversations(r.Context(), userID, pageNum, pageSize)
	if err != nil {
		writeError(w, http.StatusOK, err)
		return
	}

	items := make([]conversationItem, 0, len(messages))
	for _, m := range messages {
		items = append(items, toConversationItem(userID, m))
	}

	writeData(w, map[string]any{"items": items, "total": total})
}

func toConversationItem(userID int64, m *model.ChatMessage) conversationItem {
	peerID := m.ReceiverID
	if m.SenderID != userID {
		peerID = m.SenderID
	}
	return conversationItem{
		PeerID:         strconv.FormatInt(peerID, 10),
		RoomID:         m.RoomID,
		LastMessageID:  strconv.FormatInt(m.ID, 10),
		LastContent:    m.Content,
		LastFromUserID: strconv.FormatInt(m.SenderID, 10),
		LastAt:         m.CreatedAt.Format(time.DateTime),
	}
}
