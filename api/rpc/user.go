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

// ExchangeRefreshToken 调用用户服务换发一个新的刷新令牌。
//
// 一次调用同时完成「校验旧的、吊销旧的、落库新的」：
// 网关没有数据库连接，没法自己落库，所以不能把换发拆成两次 RPC。
// 它只在访问令牌失效时才会被调用，用访问令牌通过的那条快路径不受影响。
func ExchangeRefreshToken(ctx context.Context, userID int64, old string) (string, error) {
	resp, err := userClient.ExchangeRefreshToken(ctx, &user.ExchangeRefreshTokenRequest{
		Token:  old,
		UserId: userID,
	})
	if err != nil {
		logger.Errorf("rpc.ExchangeRefreshToken: %v", err)
		return "", errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return "", err
	}
	if resp.RefreshToken == nil || *resp.RefreshToken == "" {
		// 声称成功却没给新令牌，属于协议违约，按内部错误处理，
		// 不能让网关把空值写进响应头。
		logger.Errorf("rpc.ExchangeRefreshToken: 响应缺少 refresh_token")
		return "", errno.InternalError
	}
	return *resp.RefreshToken, nil
}
