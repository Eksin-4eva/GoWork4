package service

import (
	"encoding/json"
	"net/http"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// HandleWebSocket 处理 /ws 握手并进入消息循环。
func (s *Service) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, tokens, err := s.authenticate(r)
	if err != nil {
		// 鉴权失败不区分原因，统一 401
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resultHeader := http.Header{}
	for key, value := range tokens {
		resultHeader.Set(key, value)
	}

	conn, err := s.upgrader.Upgrade(w, r, resultHeader)
	if err != nil {
		logger.Warnf("chat: 握手失败: %v", err)
		return
	}

	sess := &session{userID: userID, conn: conn}
	s.hub.add(sess)
	defer func() {
		s.hub.remove(sess)
		_ = conn.Close()
	}()

	if err := sess.writeJSON(outgoingMessage{Type: "connected"}); err != nil {
		return
	}

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req incomingMessage
		if err := json.Unmarshal(payload, &req); err != nil {
			_ = sess.writeJSON(outgoingMessage{Type: "error", Error: "消息格式不正确"})
			continue
		}
		if err := s.handleMessage(r.Context(), userID, sess, req); err != nil {
			// 只回 errno 的文案，不回显数据库等内部错误细节
			_ = sess.writeJSON(outgoingMessage{
				Type:  "error",
				Error: errno.ConvertErr(err).ErrorMsg,
			})
		}
	}
}
