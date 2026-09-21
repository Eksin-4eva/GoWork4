package user

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// GetByUsername 按用户名查用户。第二个返回值为 false 表示不存在（不是错误）。
func (d *DBUser) GetByUsername(ctx context.Context, username string) (bool, *model.User, error) {
	u := new(model.User)

	err := d.client.WithContext(ctx).Table(constants.UserTableName).Where("username = ?", username).First(u).Error
	if err == nil {
		return true, u, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil, nil
	}
	logger.Errorf("dal.user.GetByUsername: %v", err)
	return false, nil, errno.InternalError
}
