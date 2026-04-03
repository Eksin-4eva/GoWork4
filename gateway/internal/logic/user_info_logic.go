package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic { return &UserInfoLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *UserInfoLogic) UserInfo(req *types.UserInfoReq) (resp *types.UserResp, err error) { authUser, err := l.svcCtx.App.MustAuth(l.ctx); if err != nil { return nil, err }; return l.svcCtx.App.GetUserInfo(l.ctx, authUser.UserID, req.UserId) }
