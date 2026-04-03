package logic

import (
	"context"
	"mime/multipart"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic { return &UploadAvatarLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *UploadAvatarLogic) UploadAvatar(file multipart.File, header *multipart.FileHeader) (resp *types.UserResp, err error) { authUser, err := l.svcCtx.App.MustAuth(l.ctx); if err != nil { return nil, err }; return l.svcCtx.App.UploadAvatar(l.ctx, authUser.UserID, file, header) }
