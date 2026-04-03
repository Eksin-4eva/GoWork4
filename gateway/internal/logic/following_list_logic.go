package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowingListLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewFollowingListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowingListLogic { return &FollowingListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *FollowingListLogic) FollowingList(req *types.RelationListReq) (resp *types.SocialListResp, err error) { return l.svcCtx.App.FollowingList(l.ctx, req.UserId, req.PageNum, req.PageSize) }
