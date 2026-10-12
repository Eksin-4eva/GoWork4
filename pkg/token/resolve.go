package token

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
)

// RefreshTokenExchanger 校验刷新令牌并换发一个新的。
//
// 这个动作要访问数据库（换发出来的令牌必须落库），所以由调用方把实现传进来，
// **token 包本身不依赖任何存储层**——这就是依赖倒置。
//
// 项目里有两个实现：
//   - pkg/refresh.Exchanger：连着数据库的进程（user 微服务、chat）直接用
//   - api/mw 里的小适配器：网关没有数据库连接，改为调用 user 服务的 RPC
//
// 于是同一条令牌规则，在两个进程里用两种取数方式共用。
type RefreshTokenExchanger interface {
	// ExchangeRefreshToken 用一个仍然有效的刷新令牌换发新的一个。
	// 令牌不存在 / 已吊销 / 已过期时返回 errno.AuthInvalid。
	ExchangeRefreshToken(ctx context.Context, userID int64, oldRefresh string) (string, error)
}

// ResolveUser 依次尝试访问令牌与刷新令牌，返回可信的用户 ID。
//
// 访问令牌过期时用刷新令牌换发，是「用户不用重新登录」的关键。
// 但光验签不够：还要确认刷新令牌仍在库里且没被吊销，
// 否则用户改密码或登出之后，此前签发的刷新令牌依然能用满整个有效期。
//
// 第二个返回值是**换发后的刷新令牌**：只有走了刷新令牌这条路径才有值。
// 为什么不是「只回答有效/无效」：换发必须落库，而调用方（HTTP 网关）
// 没有数据库连接，落库只能由 exchanger 那一侧完成。
//
// HTTP 网关与 WebSocket 服务共用这一份实现，避免两边规则跑偏。
func ResolveUser(
	ctx context.Context, exchanger RefreshTokenExchanger, access, refresh string,
) (userID int64, newRefresh string, err error) {
	if access != "" {
		if userID, err := CheckToken(access, constants.TypeAccessToken); err == nil {
			// 访问令牌有效：纯本地验签，不查库、不发 RPC。
			// 刷新令牌保持原样——它还在库里，客户端不需要替换。
			return userID, "", nil
		}
	}

	if refresh == "" {
		return 0, "", errno.AuthInvalid
	}

	// 先本地验签再换发：签名不对的令牌不值得浪费一次数据库查询或 RPC。
	userID, err = CheckToken(refresh, constants.TypeRefreshToken)
	if err != nil {
		return 0, "", err
	}

	newRefresh, err = exchanger.ExchangeRefreshToken(ctx, userID, refresh)
	if err != nil {
		return 0, "", err
	}
	return userID, newRefresh, nil
}
