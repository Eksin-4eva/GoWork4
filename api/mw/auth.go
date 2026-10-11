package mw

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"gobili/api/pack"
	"gobili/pkg/base"
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

		userID, err := token.ResolveUser(ctx, base.GetClientSet().DB.User, access, refresh)
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

// CurrentUserID 取出当前登录用户。只在挂了 Auth 的路由上调用才有值。
func CurrentUserID(c *app.RequestContext) int64 {
	value, ok := c.Get(constants.UserIDContextKey)
	if !ok {
		return 0
	}
	userID, _ := value.(int64)
	return userID
}
