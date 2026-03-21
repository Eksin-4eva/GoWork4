package websocket

import (
	"sync"

	"github.com/hertz-contrib/websocket"
)

// Client 定义WebSocket客户端连接
type Client struct {
	UserID int64
	Conn   *websocket.Conn
}

// Manager 定义WebSocket连接管理器
type Manager struct {
	clients    map[int64]*Client // 用户ID到客户端的映射
	rooms      map[string]map[int64]bool // 聊天室ID到用户ID的映射
	clientsMux sync.RWMutex
	roomsMux   sync.RWMutex
}

// NewManager 创建一个新的WebSocket连接管理器
func NewManager() *Manager {
	return &Manager{
		clients: make(map[int64]*Client),
		rooms:   make(map[string]map[int64]bool),
	}
}

// RegisterClient 注册一个新的客户端连接
func (m *Manager) RegisterClient(userID int64, conn *websocket.Conn) {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	m.clients[userID] = &Client{UserID: userID, Conn: conn}
}

// UnregisterClient 注销一个客户端连接
func (m *Manager) UnregisterClient(userID int64) {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	delete(m.clients, userID)
}

// GetClient 获取指定用户的客户端连接
func (m *Manager) GetClient(userID int64) (*Client, bool) {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	client, ok := m.clients[userID]
	return client, ok
}

// JoinRoom 用户加入聊天室
func (m *Manager) JoinRoom(roomID string, userID int64) {
	m.roomsMux.Lock()
	defer m.roomsMux.Unlock()
	if _, ok := m.rooms[roomID]; !ok {
		m.rooms[roomID] = make(map[int64]bool)
	}
	m.rooms[roomID][userID] = true
}

// LeaveRoom 用户离开聊天室
func (m *Manager) LeaveRoom(roomID string, userID int64) {
	m.roomsMux.Lock()
	defer m.roomsMux.Unlock()
	if users, ok := m.rooms[roomID]; ok {
		delete(users, userID)
		if len(users) == 0 {
			delete(m.rooms, roomID)
		}
	}
}

// GetRoomUsers 获取聊天室中的所有用户
func (m *Manager) GetRoomUsers(roomID string) []int64 {
	m.roomsMux.RLock()
	defer m.roomsMux.RUnlock()
	if users, ok := m.rooms[roomID]; ok {
		userIDs := make([]int64, 0, len(users))
		for userID := range users {
			userIDs = append(userIDs, userID)
		}
		return userIDs
	}
	return []int64{}
}
