// Package user 实现 Kitex 生成的 UserService 接口。
//
// 这一层只做「RPC 模型 ↔ 领域模型」的适配与错误传递，
// 业务规则全在 service 子包里。这样业务逻辑不依赖 RPC 框架，
// 将来换传输层（gRPC / HTTP）不用动它。
package user

import (
	"context"

	"gobili/internal/rpcbase"
	"gobili/internal/user/pack"
	"gobili/internal/user/service"
	"gobili/kitex_gen/user"
)

// UserServiceImpl 实现 kitex_gen/user 里的 UserService 接口。
type UserServiceImpl struct {
	svc *service.UserService
}

// NewUserServiceImpl 构造用户服务实现。
func NewUserServiceImpl() *UserServiceImpl {
	return &UserServiceImpl{svc: service.New()}
}

// Register 用户注册。
func (s *UserServiceImpl) Register(
	ctx context.Context, req *user.RegisterRequest,
) (*user.RegisterResponse, error) {
	if err := s.svc.Register(ctx, req.Username, req.Password); err != nil {
		return &user.RegisterResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &user.RegisterResponse{Base: rpcbase.OK()}, nil
}

// Login 用户登录，返回双令牌。
func (s *UserServiceImpl) Login(
	ctx context.Context, req *user.LoginRequest,
) (*user.LoginResponse, error) {
	mfaCode := ""
	if req.Code != nil {
		mfaCode = *req.Code
	}

	access, refresh, u, err := s.svc.Login(ctx, req.Username, req.Password, mfaCode)
	if err != nil {
		return &user.LoginResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &user.LoginResponse{
		Base:         rpcbase.OK(),
		User:         pack.BuildUser(u),
		AccessToken:  &access,
		RefreshToken: &refresh,
	}, nil
}

// GetUserInfo 获取用户信息。
func (s *UserServiceImpl) GetUserInfo(
	ctx context.Context, req *user.GetUserInfoRequest,
) (*user.GetUserInfoResponse, error) {
	targetID := int64(0)
	if req.TargetUserId != nil {
		targetID = *req.TargetUserId
	}

	u, err := s.svc.GetInfo(ctx, req.CurrentUserId, targetID)
	if err != nil {
		return &user.GetUserInfoResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &user.GetUserInfoResponse{Base: rpcbase.OK(), User: pack.BuildUser(u)}, nil
}

// UpdateAvatar 更新头像 key（文件由网关上传递到对象存储）。
func (s *UserServiceImpl) UpdateAvatar(
	ctx context.Context, req *user.UpdateAvatarRequest,
) (*user.UpdateAvatarResponse, error) {
	u, err := s.svc.UpdateAvatar(ctx, req.UserId, req.AvatarKey)
	if err != nil {
		return &user.UpdateAvatarResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &user.UpdateAvatarResponse{Base: rpcbase.OK(), User: pack.BuildUser(u)}, nil
}

// GetMfaQRCode 生成 MFA 绑定二维码。
func (s *UserServiceImpl) GetMfaQRCode(
	ctx context.Context, req *user.GetMfaQRCodeRequest,
) (*user.GetMfaQRCodeResponse, error) {
	secret, qrcode, err := s.svc.GetMfaQRCode(ctx, req.UserId)
	if err != nil {
		return &user.GetMfaQRCodeResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &user.GetMfaQRCodeResponse{
		Base:   rpcbase.OK(),
		Secret: &secret,
		Qrcode: &qrcode,
	}, nil
}

// BindMfa 绑定 MFA。
func (s *UserServiceImpl) BindMfa(
	ctx context.Context, req *user.BindMfaRequest,
) (*user.BindMfaResponse, error) {
	if err := s.svc.BindMfa(ctx, req.UserId, req.Secret, req.Code); err != nil {
		return &user.BindMfaResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &user.BindMfaResponse{Base: rpcbase.OK()}, nil
}

// ExchangeRefreshToken 换发刷新令牌。
//
// 网关的鉴权中间件在访问令牌失效时通过它换发新令牌——
// 微服务化之后网关不再直连数据库，「校验 + 落库」只能问用户服务。
// 把两件事合成一次调用，网关上就不会为了落库再多发一次 RPC。
func (s *UserServiceImpl) ExchangeRefreshToken(
	ctx context.Context, req *user.ExchangeRefreshTokenRequest,
) (*user.ExchangeRefreshTokenResponse, error) {
	raw, err := s.svc.ExchangeRefreshToken(ctx, req.UserId, req.Token)
	if err != nil {
		return &user.ExchangeRefreshTokenResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &user.ExchangeRefreshTokenResponse{Base: rpcbase.OK(), RefreshToken: &raw}, nil
}
