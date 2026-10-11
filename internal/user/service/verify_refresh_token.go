package service

import (
	"context"

	"gobili/pkg/errno"
)

// VerifyRefreshToken 判断刷新令牌是否仍然有效（在库且未被吊销）。
//
// 这是给网关的鉴权中间件用的：微服务化之后网关不再直连数据库，
// 「这个刷新令牌被吊销了没有」只能问用户服务。
//
// 它同时满足 pkg/token 里的 RefreshTokenChecker 接口语义，
// 但注意网关是通过 RPC 调它，不是本地调用。
func (s *UserService) VerifyRefreshToken(ctx context.Context, raw string) (bool, error) {
	if raw == "" {
		return false, errno.AuthMissing
	}
	return s.db.User.IsRefreshTokenValid(ctx, raw)
}
