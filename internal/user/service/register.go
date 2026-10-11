package service

import (
	"context"
	"strings"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// 用户名与密码的长度约束。
const (
	minUsernameLen = 3
	maxUsernameLen = 32
	minPasswordLen = 8
	maxPasswordLen = 64
)

// Register 注册新用户。
//
// 先查一次是为了给出「用户名已被占用」这种友好提示；
// 真正的唯一性由 users.uniq_username 兜底——并发注册时两个请求可能同时通过检查，
// 后到的那个会在写库时撞唯一键，由 DAL 翻译成同一个业务错误。
func (s *UserService) Register(ctx context.Context, username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return errno.ParamEmpty.WithMessage("用户名和密码不能为空")
	}
	if n := len(username); n < minUsernameLen || n > maxUsernameLen {
		return errno.ParamError.WithMessage("用户名长度需在 3 到 32 个字符之间")
	}
	if n := len(password); n < minPasswordLen || n > maxPasswordLen {
		return errno.ParamError.WithMessage("密码长度需在 8 到 64 个字符之间")
	}

	found, _, err := s.db.User.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	if found {
		return errno.BizDuplicated.WithMessage("用户名已被占用")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		logger.Errorf("user.Register: hash password: %v", err)
		return errno.InternalError
	}

	id, err := s.sf.NextVal()
	if err != nil {
		logger.Errorf("user.Register: snowflake: %v", err)
		return errno.InternalError
	}

	return s.db.User.Create(ctx, &model.User{
		ID:           id,
		Username:     username,
		PasswordHash: hash,
	})
}
