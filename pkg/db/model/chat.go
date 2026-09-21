package model

import (
	"time"

	"gorm.io/gorm"
)

// ChatMessage 对应 chat_messages 表。
//
// 对应的 DAL 在 Phase 7 迁移 chat 服务时补齐，这里先把映射关系固定下来。
type ChatMessage struct {
	ID          int64          `json:"id"`
	RoomID      string         `json:"room_id"`
	SenderID    int64          `json:"sender_id"`
	ReceiverID  int64          `json:"receiver_id"`
	MessageType string         `json:"message_type"`
	Content     string         `json:"content"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}
