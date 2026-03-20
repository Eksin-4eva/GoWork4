package model

import "time"

type Comment struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"   json:"id"`
	VideoID    int64     `gorm:"not null;index"             json:"video_id"`
	UserID     int64     `gorm:"not null"                   json:"user_id"`
	Content    string    `gorm:"type:text;not null"         json:"content"`
	ParentID   int64     `gorm:"default:0;index"            json:"parent_id"`
	LikeCount  int64     `gorm:"default:0"                  json:"like_count"`
	ChildCount int64     `gorm:"default:0"                  json:"child_count"`
	CreatedAt  time.Time `gorm:"autoCreateTime"             json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"             json:"updated_at"`
}

func (Comment) TableName() string { return "comments" }
