package service

import (
	"sync"

	"github.com/gorilla/websocket"
)

// session 是一条在线连接。
type session struct {
	userID int64
	conn   *websocket.Conn
	mu     sync.Mutex
}

// writeJSON 串行化写入。
//
// gorilla 不允许并发写同一条连接，而广播与回执可能同时发生，所以必须加锁。
func (s *session) writeJSON(v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.WriteJSON(v)
}

// hub 维护在线连接。
//
// 一个用户对应一组连接：同一个人可能同时用手机和网页登录。
type hub struct {
	mu       sync.RWMutex
	sessions map[int64]map[*session]struct{}
}

func newHub() *hub {
	return &hub{sessions: make(map[int64]map[*session]struct{})}
}

func (h *hub) add(s *session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions[s.userID] == nil {
		h.sessions[s.userID] = make(map[*session]struct{})
	}
	h.sessions[s.userID][s] = struct{}{}
}

func (h *hub) remove(s *session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	group := h.sessions[s.userID]
	delete(group, s)
	if len(group) == 0 {
		delete(h.sessions, s.userID)
	}
}

// deliver 把消息投递给某个用户的所有在线连接，返回是否至少有一个连接在线。
//
// 返回 false 不代表消息丢了：消息已经落库，对方下次拉历史还能看到。
func (h *hub) deliver(userID int64, payload any) bool {
	h.mu.RLock()
	targets := make([]*session, 0, len(h.sessions[userID]))
	for s := range h.sessions[userID] {
		targets = append(targets, s)
	}
	h.mu.RUnlock()

	if len(targets) == 0 {
		return false
	}

	for _, s := range targets {
		if err := s.writeJSON(payload); err != nil {
			// 写失败说明连接已经坏了，主动关掉让读循环尽快退出并清理
			_ = s.conn.Close()
			h.remove(s)
		}
	}
	return true
}
