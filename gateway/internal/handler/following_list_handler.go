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

func FollowingListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RelationListReq
		if err := httpx.Parse(r, &req); err != nil {
			errorResp(w, r, err)
			return
		}

		l := logic.NewFollowingListLogic(r.Context(), svcCtx)
		resp, err := l.FollowingList(&req)
		if err != nil {
			errorResp(w, r, err)
		} else {
			writeJson(w, r, resp)
		}
	}
}
