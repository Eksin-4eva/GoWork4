package user

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// IsRefreshTokenValid 判断刷新令牌是否仍然有效：
// 未吊销、未软删、且尚未过期。
func (d *DBUser) IsRefreshTokenValid(ctx context.Context, token string) (bool, error) {
	var count int64

	err := d.client.WithContext(ctx).Table(constants.RefreshTokenTableName).
		Where("token = ? AND revoked = 0 AND expires_at > NOW()", token).
		Count(&count).Error
	if err != nil {
		logger.Errorf("dal.user.IsRefreshTokenValid: %v", err)
		return false, errno.InternalError
	}
	return count > 0, nil
}
