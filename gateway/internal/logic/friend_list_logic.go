package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic { return &FriendListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *FriendListLogic) FriendList(req *types.FriendListReq) (resp *types.SocialListResp, err error) { authUser, err := l.svcCtx.App.MustAuth(l.ctx); if err != nil { return nil, err }; return l.svcCtx.App.FriendList(l.ctx, authUser.UserID, req.PageNum, req.PageSize) }
