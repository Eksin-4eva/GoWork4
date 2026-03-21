package model

import "time"

// Video 视频表
// 索引建议：user_id 普通索引（查发布列表），created_at 索引（排序），like_count 索引（热门排行）
type Video struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"              json:"id"`
	UserID       int64     `gorm:"not null;index"                        json:"user_id"`
	Title        string    `gorm:"type:varchar(256);not null"            json:"title"`
	Description  string    `gorm:"type:text"                             json:"description"`
	VideoURL     string    `gorm:"type:varchar(512);not null"            json:"video_url"`
	CoverURL     string    `gorm:"type:varchar(512)"                     json:"cover_url"`
	VisitCount   int64     `gorm:"default:0"                             json:"visit_count"`
	LikeCount    int64     `gorm:"default:0;index"                       json:"like_count"`
	CommentCount int64     `gorm:"default:0"                             json:"comment_count"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index"                  json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"                        json:"updated_at"`
}

func (Video) TableName() string { return "videos" }
