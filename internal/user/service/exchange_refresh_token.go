package service

import (
	"context"

	"gobili/pkg/refresh"
)

// ExchangeRefreshToken 校验旧刷新令牌并换发一个新的（旧的立即失效）。
//
// 这是给网关的鉴权中间件用的。微服务化之后网关不再直连数据库，
// 所以「校验 + 吊销旧的 + 落库新的」这三件事只能在用户服务里完成，
// 合并成一次 RPC —— 网关上依然只在刷新令牌这条路径上花一次调用。
//
// 换发逻辑本身在 pkg/refresh：chat 进程也要做同一件事（它直接连数据库），
// 两边共用一份实现，避免「改了一边忘了另一边」。
func (s *UserService) ExchangeRefreshToken(ctx context.Context, userID int64, old string) (string, error) {
	return refresh.Exchanger{DB: s.db, SF: s.sf}.ExchangeRefreshToken(ctx, userID, old)
}
