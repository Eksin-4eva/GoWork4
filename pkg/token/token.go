// Package token 负责双令牌的签发与校验。
//
// 放在 pkg 而不是 api/mw：登录服务要签发令牌，如果令牌逻辑住在 api/mw，
// 业务层就得反向依赖 HTTP 中间件层，依赖方向就反了。
package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gobili/config"
	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Claims 是双令牌的载荷。
//
// Type 用来区分访问令牌与刷新令牌。两者本就用不同密钥签名，
// 再叠加类型校验，是为了防止「把两个密钥配成同一个」时刷新令牌能直接当访问令牌用。
type Claims struct {
	UserID int64 `json:"user_id"`
	Type   int64 `json:"type"`
	jwt.RegisteredClaims
}

// CreateToken 按类型签发一个令牌。
func CreateToken(tokenType int64, userID int64) (string, error) {
	secret, ttl, err := secretAndTTL(tokenType)
	if err != nil {
		return "", err
	}

	now := time.Now()
	jti, err := newJti()
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    constants.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		logger.Errorf("token.CreateToken: %v", err)
		return "", errno.InternalError
	}
	return signed, nil
}

// newJti 生成令牌的唯一标识。
//
// 这个字段是必需的，不是锦上添花：JWT 的时间戳只有**秒**级精度，
// 同一个用户在同一秒内登录两次会得到完全相同的载荷 → 相同的签名 → 相同的令牌串，
// 直接撞上 refresh_tokens 表的 uniq_token 唯一键。
// 这个问题是跑端到端冒烟测试时暴露出来的。
func newJti() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("token.newJti: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// CreateAllToken 签发一对访问令牌与刷新令牌。
func CreateAllToken(userID int64) (string, string, error) {
	access, err := CreateToken(constants.TypeAccessToken, userID)
	if err != nil {
		return "", "", err
	}
	refresh, err := CreateToken(constants.TypeRefreshToken, userID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// CheckToken 校验令牌并返回其中的用户 ID。
//
// expected 是期望的令牌类型：调用方必须明确自己要的是访问令牌还是刷新令牌。
func CheckToken(raw string, expected int64) (int64, error) {
	secret, _, err := secretAndTTL(expected)
	if err != nil {
		return 0, err
	}

	claims := new(Claims)
	_, err = jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return []byte(secret), nil
	},
		// 显式限定签名算法，挡掉 alg 混淆一类的攻击
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(constants.JWTIssuer),
	)
	if err != nil {
		return 0, errno.AuthInvalid
	}
	if claims.Type != expected || claims.UserID == 0 {
		return 0, errno.AuthInvalid
	}
	return claims.UserID, nil
}

// secretAndTTL 返回指定令牌类型对应的签名密钥与有效期。
func secretAndTTL(tokenType int64) (string, time.Duration, error) {
	cfg := config.Get().JWT

	// 密钥为空时必须直接失败，否则任何字符串都能签名通过
	if cfg.AccessSecret == "" || cfg.RefreshSecret == "" {
		logger.Errorf("token: JWT 密钥未配置，拒绝签发与校验令牌")
		return "", 0, errno.InternalError
	}

	switch tokenType {
	case constants.TypeAccessToken:
		return cfg.AccessSecret, cfg.AccessTTL, nil
	case constants.TypeRefreshToken:
		return cfg.RefreshSecret, cfg.RefreshTTL, nil
	default:
		return "", 0, errno.AuthInvalid
	}
}
