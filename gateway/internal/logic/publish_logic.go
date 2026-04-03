package logic

import (
	"context"
	"mime/multipart"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic { return &PublishLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *PublishLogic) Publish(title, description string, file multipart.File, header *multipart.FileHeader) (resp *types.BaseResp, err error) { authUser, err := l.svcCtx.App.MustAuth(l.ctx); if err != nil { return nil, err }; if err := l.svcCtx.App.PublishVideo(l.ctx, authUser.UserID, title, description, file, header); err != nil { return nil, err }; return &types.BaseResp{Base: types.Base{Code: 10000, Msg: "success"}}, nil }
