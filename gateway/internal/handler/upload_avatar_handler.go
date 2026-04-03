// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"gobili/gateway/internal/logic"
	"gobili/gateway/internal/svc"
)

func UploadAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r, ok := withRequiredAuth(w, r, svcCtx)
		if !ok {
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			errorResp(w, r, err)
			return
		}
		defer file.Close()

		l := logic.NewUploadAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadAvatar(file, header)
		if err != nil {
			errorResp(w, r, err)
		} else {
			writeJson(w, r, resp)
		}
	}
}
