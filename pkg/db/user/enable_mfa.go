package user

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// EnableMfa 开启用户的 MFA，并保存 TOTP 密钥。
func (d *DBUser) EnableMfa(ctx context.Context, id int64, secret string) error {
	err := d.client.WithContext(ctx).Table(constants.UserTableName).
		Where("id = ?", id).
		Updates(map[string]any{"mfa_secret": secret, "mfa_enabled": true}).Error
	if err != nil {
		logger.Errorf("dal.user.EnableMfa: %v", err)
		return errno.InternalError
	}
	return nil
}
