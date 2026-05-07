package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchVideoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchVideoLogic {
	return &SearchVideoLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *SearchVideoLogic) SearchVideo(req *types.SearchVideoReq) (resp *types.PagedVideoListResp, err error) {
	return l.svcCtx.App.SearchVideos(l.ctx, req)
}
