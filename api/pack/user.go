package pack

import (
	hzmodel "gobili/api/model/model"
	dbmodel "gobili/pkg/db/model"
)

// BuildUser 把数据库模型转成对外的用户模型。
func BuildUser(u *dbmodel.User) *hzmodel.User {
	if u == nil {
		return nil
	}
	return &hzmodel.User{
		ID:        FormatID(u.ID),
		Username:  u.Username,
		AvatarURL: ObjectURL(u.AvatarKey),
		CreatedAt: FormatTime(u.CreatedAt),
		UpdatedAt: FormatTime(u.UpdatedAt),
	}
}
