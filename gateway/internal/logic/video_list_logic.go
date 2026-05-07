package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VideoListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoListLogic {
	return &VideoListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *VideoListLogic) VideoList(req *types.PublishListReq) (resp *types.PagedVideoListResp, err error) {
	return l.svcCtx.App.VideoList(l.ctx, req.UserId, req.PageNum, req.PageSize)
}
