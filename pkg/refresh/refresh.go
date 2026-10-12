// Package refresh 实现「刷新令牌换发」这一个动作。
//
// 换发必须同时做三件事：签发新令牌、让旧令牌立即失效、把新令牌落库。
// 少做最后一件就会出事——客户端按约定替换掉本地令牌之后，这次替换根本没被记录，
// 下次换发查库查不到，用户被迫重新登录。
//
// 为什么放在 pkg 而不是某个服务的私有目录：两个进程要做同一件事，
//   - user 微服务（网关通过 RPC 调它）
//   - chat 进程（本地直接调用，它自己连着数据库）
//
// 而 token 包不能承担这件事——它不认识数据库，那正是依赖倒置要守住的方向。
package refresh

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/db"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/token"
	"gobili/pkg/utils"
)

// Exchanger 用数据库里的一次原子替换完成刷新令牌换发。
//
// 它满足 pkg/token 的 RefreshTokenExchanger 接口——
// **token 包不知道换发背后是本地数据库还是跨进程 RPC**。
type Exchanger struct {
	DB *db.Database
	SF *utils.Snowflake
}

// ExchangeRefreshToken 校验旧刷新令牌并换发一个新的（旧的立即失效）。
//
// 旧令牌不存在、已吊销、已过期或不属于该用户时返回 errno.AuthInvalid。
func (e Exchanger) ExchangeRefreshToken(ctx context.Context, userID int64, old string) (string, error) {
	if old == "" || userID <= 0 {
		return "", errno.AuthInvalid
	}

	// 先签发，再落库：反过来的话，落库成功但签发失败会留下一条永远用不了的记录。
	raw, expiresAt, err := token.CreateTokenWithExpiry(constants.TypeRefreshToken, userID)
	if err != nil {
		return "", err
	}

	tokenID, err := e.SF.NextVal()
	if err != nil {
		logger.Errorf("refresh.ExchangeRefreshToken: snowflake: %v", err)
		return "", errno.InternalError
	}

	swapped, err := e.DB.User.ReplaceRefreshToken(ctx, old, userID, &model.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		Token:     raw,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}
	if !swapped {
		return "", errno.AuthInvalid
	}
	return raw, nil
}
