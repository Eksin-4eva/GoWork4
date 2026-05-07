package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowerListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFollowerListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowerListLogic {
	return &FollowerListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *FollowerListLogic) FollowerList(req *types.RelationListReq) (resp *types.SocialListResp, err error) {
	return l.svcCtx.App.FollowerList(l.ctx, req.UserId, req.PageNum, req.PageSize)
}
