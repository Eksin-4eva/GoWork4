package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeListLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewLikeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeListLogic { return &LikeListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *LikeListLogic) LikeList(req *types.LikeListReq) (resp *types.VideoListResp, err error) { return l.svcCtx.App.LikeList(l.ctx, req.UserId, req.PageNum, req.PageSize) }
