package logic

import (
	"context"

	"gobili/gateway/internal/svc"
	"gobili/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMfaQrcodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMfaQrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMfaQrcodeLogic {
	return &GetMfaQrcodeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *GetMfaQrcodeLogic) GetMfaQrcode() (resp *types.MfaQRCodeResp, err error) {
	authUser, err := l.svcCtx.App.MustAuth(l.ctx)
	if err != nil {
		return nil, err
	}
	return l.svcCtx.App.GenerateMFA(l.ctx, authUser.UserID)
}
