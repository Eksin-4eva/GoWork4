package model

import "time"

type CommentLike struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_user_comment" json:"user_id"`
	CommentID int64     `gorm:"not null;uniqueIndex:idx_user_comment" json:"comment_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (CommentLike) TableName() string { return "comment_likes" }
