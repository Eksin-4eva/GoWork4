// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"gobili/gateway/internal/logic"
	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			errorResp(w, r, err)
			return
		}

		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, tokens, err := l.Login(&req)
		if err != nil {
			errorResp(w, r, err)
			return
		}
		for key, value := range tokens {
			if value != "" {
				w.Header().Set(key, value)
			}
		}
		writeJson(w, r, resp)
	}
}
