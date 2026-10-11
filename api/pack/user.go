package pack

import (
	hzmodel "gobili/api/model/model"
	"gobili/kitex_gen/rpcmodel"
)

// BuildUser 把 RPC 返回的用户模型转成对外的 HTTP 模型。
func BuildUser(u *rpcmodel.User) *hzmodel.User {
	if u == nil {
		return nil
	}
	return &hzmodel.User{
		ID:        FormatID(u.Id),
		Username:  u.Username,
		AvatarURL: ObjectURL(u.AvatarKey), // ← key 在这里变成可访问的 URL
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
