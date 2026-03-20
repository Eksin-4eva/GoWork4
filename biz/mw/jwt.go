package mw

import (
	"context"
	"net/http"
	"os"

	"github.com/BiliGO/biz/pkg/response"
	"github.com/BiliGO/pkg/utils"
	"github.com/cloudwego/hertz/pkg/app"
)

const (
	CtxUserIDKey = "user_id"
)

// JWTAuth 验证 Access Token，过期但 Refresh Token 有效时返回特定错误码
func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		secret := os.Getenv("JWT_SECRET")

		accessToken := string(c.GetHeader("Access_Token"))
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, response.Fail(response.CodeTokenInvalid, "missing access token"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(accessToken, secret)
		if err == utils.ErrTokenExpired {
			refreshToken := string(c.GetHeader("Refresh_Token"))
			if refreshToken == "" {
				c.JSON(http.StatusUnauthorized, response.Fail(response.CodeAccessExpired, "access token expired, please refresh"))
				c.Abort()
				return
			}
			rClaims, rErr := utils.ParseToken(refreshToken, secret)
			if rErr != nil || rClaims.TokenType != utils.TokenTypeRefresh {
				c.JSON(http.StatusUnauthorized, response.Fail(response.CodeAccessExpired, "access token expired, refresh token invalid"))
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, response.Fail(response.CodeAccessExpired, "access token expired, please use refresh token to get new tokens"))
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Fail(response.CodeTokenInvalid, "invalid token"))
			c.Abort()
			return
		}
		if claims.TokenType != utils.TokenTypeAccess {
			c.JSON(http.StatusUnauthorized, response.Fail(response.CodeTokenInvalid, "wrong token type"))
			c.Abort()
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Next(ctx)
	}
}
