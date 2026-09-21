package model

import (
	"time"

	"gorm.io/gorm"
)

// Video 对应 videos 表。
//
// VideoKey / CoverKey 存的是 MinIO 的 object key（如 video/123.mp4），
// 不是完整 URL。对外 URL 由 api/pack 用 config.MinIO.PublicBaseURL 拼出来，
// 这样换存储或换域名都不需要刷库。
type Video struct {
	ID           int64          `json:"id"`
	UserID       int64          `json:"user_id"`
	VideoKey     string         `json:"video_key"`
	CoverKey     string         `json:"cover_key"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	VisitCount   int64          `json:"visit_count"`
	CommentCount int64          `json:"comment_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty"`
}
