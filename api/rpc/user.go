package rpc

import (
	"context"

	"gobili/internal/rpcbase"
	"gobili/kitex_gen/user"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Register 调用用户服务注册。
func Register(ctx context.Context, req *user.RegisterRequest) error {
	resp, err := userClient.Register(ctx, req)
	if err != nil {
		logger.Errorf("rpc.Register: %v", err)
		return errno.InternalError
	}
	return rpcbase.Check(resp.Base)
}

// Login 调用用户服务登录。
func Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	resp, err := userClient.Login(ctx, req)
	if err != nil {
		logger.Errorf("rpc.Login: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUserInfo 调用用户服务获取用户信息。
func GetUserInfo(ctx context.Context, req *user.GetUserInfoRequest) (*user.GetUserInfoResponse, error) {
	resp, err := userClient.GetUserInfo(ctx, req)
	if err != nil {
		logger.Errorf("rpc.GetUserInfo: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateAvatar 调用用户服务更新头像 key。
func UpdateAvatar(ctx context.Context, req *user.UpdateAvatarRequest) (*user.UpdateAvatarResponse, error) {
	resp, err := userClient.UpdateAvatar(ctx, req)
	if err != nil {
		logger.Errorf("rpc.UpdateAvatar: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetMfaQRCode 调用用户服务获取 MFA 二维码。
func GetMfaQRCode(ctx context.Context, req *user.GetMfaQRCodeRequest) (*user.GetMfaQRCodeResponse, error) {
	resp, err := userClient.GetMfaQRCode(ctx, req)
	if err != nil {
		logger.Errorf("rpc.GetMfaQRCode: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// BindMfa 调用用户服务绑定 MFA。
func BindMfa(ctx context.Context, req *user.BindMfaRequest) error {
	resp, err := userClient.BindMfa(ctx, req)
	if err != nil {
		logger.Errorf("rpc.BindMfa: %v", err)
		return errno.InternalError
	}
	return rpcbase.Check(resp.Base)
}

// VerifyRefreshToken 调用用户服务校验刷新令牌。
//
// 网关的鉴权中间件靠它判断「这个刷新令牌被吊销了没有」。
func VerifyRefreshToken(ctx context.Context, token string) (bool, error) {
	resp, err := userClient.VerifyRefreshToken(ctx, &user.VerifyRefreshTokenRequest{Token: token})
	if err != nil {
		logger.Errorf("rpc.VerifyRefreshToken: %v", err)
		return false, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return false, err
	}
	return resp.Valid != nil && *resp.Valid, nil
}
