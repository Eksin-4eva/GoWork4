package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FeedLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewFeedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FeedLogic { return &FeedLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *FeedLogic) Feed(req *types.FeedReq) (resp *types.VideoListResp, err error) { return l.svcCtx.App.Feed(l.ctx, req.LatestTime) }
