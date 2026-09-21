package user

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// UpdateAvatarKey 更新用户头像的 object key。
func (d *DBUser) UpdateAvatarKey(ctx context.Context, id int64, key string) error {
	err := d.client.WithContext(ctx).Table(constants.UserTableName).
		Where("id = ?", id).
		Update("avatar_key", key).Error
	if err != nil {
		logger.Errorf("dal.user.UpdateAvatarKey: %v", err)
		return errno.InternalError
	}
	return nil
}
