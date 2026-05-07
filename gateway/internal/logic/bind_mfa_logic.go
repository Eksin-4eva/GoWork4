package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindMfaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindMfaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindMfaLogic {
	return &BindMfaLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *BindMfaLogic) BindMfa(req *types.BindMfaReq) (resp *types.BaseResp, err error) {
	authUser, err := l.svcCtx.App.MustAuth(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.App.BindMFA(l.ctx, authUser.UserID, req.Secret, req.Code); err != nil {
		return nil, err
	}
	return &types.BaseResp{Base: types.Base{Code: 10000, Msg: "success"}}, nil
}
