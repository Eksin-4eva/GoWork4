package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentReq) (resp *types.BaseResp, err error) {
	authUser, err := l.svcCtx.App.MustAuth(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.App.DeleteComment(l.ctx, authUser.UserID, req.CommentId); err != nil {
		return nil, err
	}
	return &types.BaseResp{Base: types.Base{Code: 10000, Msg: "success"}}, nil
}
