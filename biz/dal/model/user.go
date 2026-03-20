package model

import "time"

// User 用户表
// 索引建议：username 唯一索引
type User struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"                json:"id"`
	Username       string    `gorm:"type:varchar(64);uniqueIndex;not null"   json:"username"`
	PasswordHash   string    `gorm:"type:varchar(256);not null"              json:"-"`
	AvatarURL      string    `gorm:"type:varchar(512)"                       json:"avatar_url"`
	MFASecret      string    `gorm:"type:varchar(64)"                        json:"-"`
	FollowerCount  int64     `gorm:"default:0"                               json:"follower_count"`
	FollowingCount int64     `gorm:"default:0"                               json:"following_count"`
	CreatedAt      time.Time `gorm:"autoCreateTime"                          json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"                          json:"updated_at"`
}

func (User) TableName() string { return "users" }
