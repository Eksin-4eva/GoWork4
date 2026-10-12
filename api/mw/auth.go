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
// 同时在每次请求成功时换发新的访问令牌；如果这次是**靠刷新令牌**通过的，
// 再把换发后的刷新令牌一并下发（客户端应替换本地存储）。这样做的收益是：
// 令牌的有效期是「最后一次请求 + TTL」，活跃用户不会突然掉线，
// 长期不用的用户则会自然失效。
//
// 注意刷新令牌只在它被用到时才换发。反过来（每次请求都下发新刷新令牌）
// 会有一个致命问题：**网关没有数据库连接，落不了库**——
// 客户端按约定替换之后，2 小时后再来换发就会查不到记录，直接被踢下线。
// 换发动作落在 user 服务（RPC）与 chat 进程（本地）里，见 pkg/refresh。
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

		userID, newRefresh, err := token.ResolveUser(ctx, refreshTokenExchanger{}, access, refresh)
		if err != nil {
			pack.RespError(c, err)
			c.Abort()
			return
		}

		newAccess, err := token.CreateToken(constants.TypeAccessToken, userID)
		if err != nil {
			pack.RespError(c, err)
			c.Abort()
			return
		}
		c.Header(constants.AccessTokenHeader, newAccess)
		if newRefresh != "" {
			c.Header(constants.RefreshTokenHeader, newRefresh)
		}

		c.Set(constants.UserIDContextKey, userID)
		c.Next(ctx)
	}
}

// refreshTokenExchanger 通过 RPC 完成刷新令牌换发。
//
// 微服务化之后网关不再直连数据库，所以「校验 + 落库」只能问用户服务。
// 它满足 pkg/token 的 RefreshTokenExchanger 接口——
// **token 包不需要知道背后是本地调用还是跨进程 RPC**，
// 这正是「依赖抽象而不是实现」的价值。
type refreshTokenExchanger struct{}

func (refreshTokenExchanger) ExchangeRefreshToken(
	ctx context.Context, userID int64, old string,
) (string, error) {
	return rpc.ExchangeRefreshToken(ctx, userID, old)
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
