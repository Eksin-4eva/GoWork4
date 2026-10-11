package token

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
)

// RefreshTokenChecker 判断刷新令牌是否仍然有效（在库且未被吊销）。
//
// 这个判断要访问数据库，所以由调用方把 DAL 传进来，
// token 包本身不依赖任何存储层。
type RefreshTokenChecker interface {
	IsRefreshTokenValid(ctx context.Context, token string) (bool, error)
}

// ResolveUser 依次尝试访问令牌与刷新令牌，返回可信的用户 ID。
//
// 访问令牌过期时用刷新令牌换发，是「用户不用重新登录」的关键。
// 但光验签不够：还要确认刷新令牌仍在库里且没被吊销，
// 否则用户改密码或登出之后，此前签发的刷新令牌依然能用满整个有效期。
//
// HTTP 网关与 WebSocket 服务共用这一份实现，避免两边规则跑偏。
func ResolveUser(
	ctx context.Context, checker RefreshTokenChecker, access, refresh string,
) (int64, error) {
	if access != "" {
		if userID, err := CheckToken(access, constants.TypeAccessToken); err == nil {
			return userID, nil
		}
	}

	if refresh == "" {
		return 0, errno.AuthInvalid
	}
	userID, err := CheckToken(refresh, constants.TypeRefreshToken)
	if err != nil {
		return 0, err
	}

	valid, err := checker.IsRefreshTokenValid(ctx, refresh)
	if err != nil {
		return 0, err
	}
	if !valid {
		return 0, errno.AuthInvalid
	}
	return userID, nil
}
