package handler

import (
	"context"
	"net/http"

	"gobili/gateway/internal/app"
	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func withOptionalAuth(r *http.Request, svcCtx *svc.ServiceContext) *http.Request {
	accessToken := r.Header.Get(svcCtx.Config.Auth.AccessHeader)
	refreshToken := r.Header.Get(svcCtx.Config.Auth.RefreshHeader)
	authUser, err := svcCtx.App.ParseUserFromTokens(r.Context(), accessToken, refreshToken)
	if err != nil {
		return r
	}
	for key, value := range authUser.Tokens {
		if value != "" {
			r.Header.Set(key, value)
		}
	}
	ctx := context.WithValue(r.Context(), "authUser", app.AuthUser{UserID: authUser.UserID, Tokens: authUser.Tokens})
	return r.WithContext(ctx)
}

func withRequiredAuth(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext) (*http.Request, bool) {
	r = withOptionalAuth(r, svcCtx)
	if _, err := svcCtx.App.MustAuth(r.Context()); err != nil {
		httpx.OkJsonCtx(r.Context(), w, types.BaseResp{Base: types.Base{Code: 10002, Msg: err.Error()}})
		return nil, false
	}
	return r, true
}

func errorResp(w http.ResponseWriter, r *http.Request, err error) {
	httpx.OkJsonCtx(r.Context(), w, types.BaseResp{Base: types.Base{Code: 10001, Msg: err.Error()}})
}

func writeJson(w http.ResponseWriter, r *http.Request, resp any) {
	httpx.OkJsonCtx(r.Context(), w, resp)
}
