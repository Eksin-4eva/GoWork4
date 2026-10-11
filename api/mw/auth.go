package mw

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"gobili/api/pack"
	"gobili/api/rpc"
	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/token"
)

// Auth 校验令牌，通过后把当前用户 ID 写入请求上下文。
//
// 同时在每次请求成功时轮换双令牌并通过响应头下发，客户端拿到后应替换本地存储。
// 这样做的收益是：令牌的有效期是「最后一次请求 + TTL」，
// 活跃用户不会突然掉线，长期不用的用户则会自然失效。
//
// 校验失败一律返回同一个错误码，不区分「没带」「过期」「类型不对」，
// 避免给攻击者提供额外的判断依据。
func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		access := string(c.GetHeader(constants.AuthHeader))
		refresh := string(c.GetHeader(constants.RefreshTokenHeader))

		if access == "" && refresh == "" {
			pack.RespError(c, errno.AuthMissing)
			c.Abort()
			return
		}

		userID, err := token.ResolveUser(ctx, refreshTokenChecker{}, access, refresh)
		if err != nil {
			pack.RespError(c, err)
			c.Abort()
			return
		}

		newAccess, newRefresh, err := token.CreateAllToken(userID)
		if err != nil {
			pack.RespError(c, err)
			c.Abort()
			return
		}
		c.Header(constants.AccessTokenHeader, newAccess)
		c.Header(constants.RefreshTokenHeader, newRefresh)

		c.Set(constants.UserIDContextKey, userID)
		c.Next(ctx)
	}
}

// refreshTokenChecker 通过 RPC 检查刷新令牌是否已被吊销。
//
// 微服务化之后网关不再直连数据库，所以这个判断只能问用户服务。
// 它满足 pkg/token 的 RefreshTokenChecker 接口——
// **token 包不需要知道背后是本地调用还是跨进程 RPC**，
// 这正是「依赖抽象而不是实现」的价值。
type refreshTokenChecker struct{}

func (refreshTokenChecker) IsRefreshTokenValid(ctx context.Context, raw string) (bool, error) {
	return rpc.VerifyRefreshToken(ctx, raw)
}

// CurrentUserID 取出当前登录用户。只在挂了 Auth 的路由上调用才有值。
func CurrentUserID(c *app.RequestContext) int64 {
	value, ok := c.Get(constants.UserIDContextKey)
	if !ok {
		return 0
	}
	userID, _ := value.(int64)
	return userID
}
