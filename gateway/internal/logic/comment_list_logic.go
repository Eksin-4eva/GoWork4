package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentListLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentListLogic { return &CommentListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *CommentListLogic) CommentList(req *types.CommentListReq) (resp *types.CommentListResp, err error) { return l.svcCtx.App.CommentList(l.ctx, req) }
