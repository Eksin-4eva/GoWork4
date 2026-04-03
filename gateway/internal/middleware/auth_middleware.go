package middleware

import (
	"context"
	"net/http"

	"gobili/gateway/internal/app"
	"gobili/gateway/internal/svc"
)

type AuthMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewAuthMiddleware(svcCtx *svc.ServiceContext) *AuthMiddleware {
	return &AuthMiddleware{svcCtx: svcCtx}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get(m.svcCtx.Config.Auth.AccessHeader)
		refreshToken := r.Header.Get(m.svcCtx.Config.Auth.RefreshHeader)
		authUser, err := m.svcCtx.App.ParseUserFromTokens(r.Context(), accessToken, refreshToken)
		if err == nil {
			for key, value := range authUser.Tokens {
				if value != "" {
					w.Header().Set(key, value)
				}
			}
			ctx := context.WithValue(r.Context(), "authUser", app.AuthUser{UserID: authUser.UserID, Tokens: authUser.Tokens})
			next(w, r.WithContext(ctx))
			return
		}
		next(w, r)
	}
}
