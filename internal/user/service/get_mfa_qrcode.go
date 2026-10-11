package service

import (
	"context"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// GetMfaQRCode 生成一个新的 TOTP 密钥，并返回可扫码的二维码。
//
// 此时**不落库**：必须等用户拿验证码确认绑定成功（BindMfa）才写入。
// 否则用户中途放弃，账号就会卡在一个他并不知道密钥的状态。
func (s *UserService) GetMfaQRCode(ctx context.Context, userID int64) (secret, qrcode string, err error) {
	found, u, err := s.db.User.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if !found {
		return "", "", errno.BizNotExist.WithMessage("用户不存在")
	}

	secret, err = utils.GenerateMfaSecret()
	if err != nil {
		logger.Errorf("user.GetMfaQRCode: generate secret: %v", err)
		return "", "", errno.InternalError
	}

	qrcode, err = utils.QRCodeDataURI(utils.MfaQRCodeURI(u.Username, secret))
	if err != nil {
		logger.Errorf("user.GetMfaQRCode: render qrcode: %v", err)
		return "", "", errno.InternalError
	}
	return secret, qrcode, nil
}
