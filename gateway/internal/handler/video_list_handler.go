// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobili/gateway/internal/logic"
	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"
)

func VideoListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublishListReq
		if err := httpx.Parse(r, &req); err != nil {
			errorResp(w, r, err)
			return
		}

		l := logic.NewVideoListLogic(r.Context(), svcCtx)
		resp, err := l.VideoList(&req)
		if err != nil {
			errorResp(w, r, err)
		} else {
			writeJson(w, r, resp)
		}
	}
}
