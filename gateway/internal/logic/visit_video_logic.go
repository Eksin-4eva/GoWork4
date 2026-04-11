package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VisitVideoLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewVisitVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VisitVideoLogic { return &VisitVideoLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *VisitVideoLogic) VisitVideo(req *types.VisitVideoReq) (resp *types.BaseResp, err error) { authUser, err := l.svcCtx.App.MustAuth(l.ctx); if err != nil { return nil, err }; if err := l.svcCtx.App.RecordVideoVisit(l.ctx, authUser.UserID, req.VideoId); err != nil { return nil, err }; return &types.BaseResp{Base: types.Base{Code: 10000, Msg: "success"}}, nil }
