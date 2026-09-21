package user

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// SaveRefreshToken 把刷新令牌落库，供后续校验与吊销。
func (d *DBUser) SaveRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	err := d.client.WithContext(ctx).Table(constants.RefreshTokenTableName).Create(token).Error
	if err != nil {
		logger.Errorf("dal.user.SaveRefreshToken: %v", err)
		return errno.InternalError
	}
	return nil
}
