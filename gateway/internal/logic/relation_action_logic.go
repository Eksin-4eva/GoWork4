package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RelationActionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRelationActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RelationActionLogic {
	return &RelationActionLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *RelationActionLogic) RelationAction(req *types.RelationActionReq) (resp *types.BaseResp, err error) {
	authUser, err := l.svcCtx.App.MustAuth(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.App.RelationAction(l.ctx, authUser.UserID, req); err != nil {
		return nil, err
	}
	return &types.BaseResp{Base: types.Base{Code: 10000, Msg: "success"}}, nil
}
