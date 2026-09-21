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

// GetByID 按主键查用户。第二个返回值为 false 表示不存在（不是错误）。
func (d *DBUser) GetByID(ctx context.Context, id int64) (bool, *model.User, error) {
	u := new(model.User)

	err := d.client.WithContext(ctx).Table(constants.UserTableName).Where("id = ?", id).First(u).Error
	if err == nil {
		return true, u, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil, nil
	}
	logger.Errorf("dal.user.GetByID: %v", err)
	return false, nil, errno.InternalError
}
