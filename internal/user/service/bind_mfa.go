package service

import (
	"context"

	"gobili/pkg/errno"
	"gobili/pkg/utils"
)

// BindMfa 校验验证码后开启 MFA。
//
// 密钥是上一步下发的、由客户端带回来的，这里必须先验码再落库。
func (s *UserService) BindMfa(ctx context.Context, userID int64, secret, code string) error {
	if secret == "" || code == "" {
		return errno.ParamEmpty.WithMessage("密钥与验证码不能为空")
	}
	if !utils.VerifyMfaCode(secret, code) {
		return errno.BizError.WithMessage("验证码错误或已过期")
	}
	return s.db.User.EnableMfa(ctx, userID, secret)
}
