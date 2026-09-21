// Package model 定义与 config/sql/init.sql 一一对应的 GORM 模型。
//
// 这里不使用 AutoMigrate，模型只是查询映射，改结构必须先改 DDL。
// 字段名依赖 GORM 默认的 snake_case 转换，注意：
//   - ID / UserID 这类后缀会被正确转成 id / user_id
//   - Mfa 不是 Go 常见缩写，所以字段名写成 MfaSecret 而非 MFASecret，
//     否则会映射成一个不存在的列名
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 对应 users 表。
type User struct {
	ID           int64          `json:"id"`
	Username     string         `json:"username"`
	PasswordHash string         `json:"password_hash"`
	AvatarKey    string         `json:"avatar_key"`
	MfaSecret    string         `json:"mfa_secret"`
	MfaEnabled   bool           `json:"mfa_enabled"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty"`
}

// RefreshToken 对应 refresh_tokens 表。
type RefreshToken struct {
	ID        int64          `json:"id"`
	UserID    int64          `json:"user_id"`
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	Revoked   bool           `json:"revoked"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}
