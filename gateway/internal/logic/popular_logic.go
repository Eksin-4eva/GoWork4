package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PopularLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewPopularLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PopularLogic { return &PopularLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *PopularLogic) Popular(req *types.PopularReq) (resp *types.VideoListResp, err error) { return l.svcCtx.App.Popular(l.ctx, req.PageNum, req.PageSize) }
