package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 对应 comments 表。
//
// ParentID 为 0 表示顶层评论。绝不要用空字符串表示"没有父评论"——
// 旧版就是这么写的，空串写进 bigint 列在严格模式下会直接报 1366。
type Comment struct {
	ID         int64          `json:"id"`
	UserID     int64          `json:"user_id"`
	VideoID    int64          `json:"video_id"`
	ParentID   int64          `json:"parent_id"`
	Content    string         `json:"content"`
	ChildCount int64          `json:"child_count"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty"`
}
