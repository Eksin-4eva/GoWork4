package service

import (
	"context"
	"strings"
	"time"

	"gobili/config"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/token"
	"gobili/pkg/utils"
)

// Login 校验账号密码并签发一对令牌。
//
// 「用户不存在」与「密码错误」返回同一句话：否则这个接口就成了枚举用户名的工具。
func (s *UserService) Login(
	ctx context.Context, username, password, mfaCode string,
) (access, refresh string, user *model.User, err error) {
	found, u, err := s.db.User.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return "", "", nil, err
	}
	if !found || !utils.CheckPassword(u.PasswordHash, password) {
		return "", "", nil, errno.BizError.WithMessage("用户名或密码错误")
	}

	if u.MfaEnabled && !utils.VerifyMfaCode(u.MfaSecret, mfaCode) {
		return "", "", nil, errno.BizError.WithMessage("验证码错误或已过期")
	}

	access, refresh, err = token.CreateAllToken(u.ID)
	if err != nil {
		return "", "", nil, err
	}

	// 刷新令牌落库后才能被吊销（登出、改密码），
	// 中间件也会回查这张表，光验签是不够的。
	tokenID, err := s.sf.NextVal()
	if err != nil {
		logger.Errorf("user.Login: snowflake: %v", err)
		return "", "", nil, errno.InternalError
	}
	err = s.db.User.SaveRefreshToken(ctx, &model.RefreshToken{
		ID:        tokenID,
		UserID:    u.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(config.Get().JWT.RefreshTTL),
	})
	if err != nil {
		return "", "", nil, err
	}
	return access, refresh, u, nil
}
