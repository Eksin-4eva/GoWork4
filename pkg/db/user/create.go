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

// Create 插入一条用户记录。
//
// 用户名冲突由 users.uniq_username 兜底：GORM 开启了 TranslateError，
// 所以这里可以直接判 gorm.ErrDuplicatedKey，不必去匹配错误字符串。
func (d *DBUser) Create(ctx context.Context, u *model.User) error {
	err := d.client.WithContext(ctx).Table(constants.UserTableName).Create(u).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return errno.BizDuplicated.WithMessage("用户名已被占用")
	}
	logger.Errorf("dal.user.Create: %v", err)
	return errno.InternalError
}
