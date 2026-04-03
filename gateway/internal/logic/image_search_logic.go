package logic

import (
	"context"
	"mime/multipart"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ImageSearchLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewImageSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImageSearchLogic { return &ImageSearchLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *ImageSearchLogic) ImageSearch(file multipart.File, header *multipart.FileHeader) (resp *types.ImageSearchResp, err error) { authUser, _ := l.svcCtx.App.MustAuth(l.ctx); return l.svcCtx.App.ImageSearch(l.ctx, authUser.UserID, file, header) }
