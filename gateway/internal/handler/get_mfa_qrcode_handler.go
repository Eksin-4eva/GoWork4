// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"gobili/gateway/internal/logic"
	"gobili/gateway/internal/svc"
)

func GetMfaQrcodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r, ok := withRequiredAuth(w, r, svcCtx)
		if !ok {
			return
		}
		l := logic.NewGetMfaQrcodeLogic(r.Context(), svcCtx)
		resp, err := l.GetMfaQrcode()
		if err != nil {
			errorResp(w, r, err)
		} else {
			writeJson(w, r, resp)
		}
	}
}
