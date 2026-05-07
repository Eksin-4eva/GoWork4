package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentPublishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentPublishLogic {
	return &CommentPublishLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *CommentPublishLogic) CommentPublish(req *types.CommentPublishReq) (resp *types.BaseResp, err error) {
	authUser, err := l.svcCtx.App.MustAuth(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.App.PublishComment(l.ctx, authUser.UserID, req); err != nil {
		return nil, err
	}
	return &types.BaseResp{Base: types.Base{Code: 10000, Msg: "success"}}, nil
}
