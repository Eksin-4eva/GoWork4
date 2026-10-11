// Package pack 负责「数据库模型 → RPC 模型」的转换。
//
// 注意这里**不做 key → URL 的转换**：对外 URL 取决于客户端从哪个域名访问，
// 那是网关该管的事。RPC 层只传 object key。
package pack

import (
	"time"

	"gobili/kitex_gen/rpcmodel"
	dbmodel "gobili/pkg/db/model"
)

// timeLayout 是 RPC 层传递时间的统一格式。
const timeLayout = "2006-01-02 15:04:05"

// BuildUser 把数据库模型转成 RPC 模型。
func BuildUser(u *dbmodel.User) *rpcmodel.User {
	if u == nil {
		return nil
	}
	return &rpcmodel.User{
		Id:        u.ID,
		Username:  u.Username,
		AvatarKey: u.AvatarKey,
		CreatedAt: formatTime(u.CreatedAt),
		UpdatedAt: formatTime(u.UpdatedAt),
	}
}

// formatTime 零值返回空串，避免把 "0001-01-01 00:00:00" 传给上层。
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(timeLayout)
}
