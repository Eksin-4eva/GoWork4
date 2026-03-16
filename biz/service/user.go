package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/BiliGO/biz/dal/model"
	"github.com/BiliGO/biz/dal/mysql"
	"github.com/BiliGO/biz/dal/query"
	"github.com/BiliGO/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterResult 注册结果
type RegisterResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

// LoginResult 登录结果
type LoginResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

// UserInfoResult 用户信息结果
type UserInfoResult struct {
	ID             string
	Username       string
	AvatarURL      string
	FollowerCount  int64
	FollowingCount int64
	CreatedAt      string
	UpdatedAt      string
}

// Register 注册用户
func Register(ctx context.Context, username, password string) (*RegisterResult, error) {
	q := query.Use(mysql.DB)
	u := q.User

	// 检查用户名是否已存在
	_, err := u.WithContext(ctx).Where(u.Username.Eq(username)).First()
	if err == nil {
		return nil, errors.New("username already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("db error: %w", err)
	}

	// 密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := u.WithContext(ctx).Create(user); err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	secret := os.Getenv("JWT_SECRET")
	accessToken, err := utils.GenerateAccessToken(user.ID, secret)
	if err != nil {
		return nil, err
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID, secret)
	if err != nil {
		return nil, err
	}

	return &RegisterResult{
		UserID:       strconv.FormatInt(user.ID, 10),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login 用户登录
func Login(ctx context.Context, username, password string) (*LoginResult, error) {
	q := query.Use(mysql.DB)
	u := q.User

	user, err := u.WithContext(ctx).Where(u.Username.Eq(username)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("wrong password")
	}

	secret := os.Getenv("JWT_SECRET")
	accessToken, err := utils.GenerateAccessToken(user.ID, secret)
	if err != nil {
		return nil, err
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID, secret)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:       strconv.FormatInt(user.ID, 10),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserInfo 获取用户信息
func GetUserInfo(ctx context.Context, userID int64) (*UserInfoResult, error) {
	q := query.Use(mysql.DB)
	u := q.User

	user, err := u.WithContext(ctx).Where(u.ID.Eq(userID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return &UserInfoResult{
		ID:             strconv.FormatInt(user.ID, 10),
		Username:       user.Username,
		AvatarURL:      user.AvatarURL,
		FollowerCount:  user.FollowerCount,
		FollowingCount: user.FollowingCount,
		CreatedAt:      user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateAvatar 更新用户头像
func UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error {
	q := query.Use(mysql.DB)
	u := q.User

	_, err := u.WithContext(ctx).Where(u.ID.Eq(userID)).UpdateSimple(u.AvatarURL.Value(avatarURL))
	return err
}

// ParseRefreshToken 解析并验证 refresh token
func ParseRefreshToken(tokenStr, secret string) (*utils.Claims, error) {
	claims, err := utils.ParseToken(tokenStr, secret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != utils.TokenTypeRefresh {
		return nil, utils.ErrTokenInvalid
	}
	return claims, nil
}

// RefreshTokens 用 refresh token 换取新的双 token
func RefreshTokens(ctx context.Context, userID int64) (accessToken, refreshToken string, err error) {
	secret := os.Getenv("JWT_SECRET")
	accessToken, err = utils.GenerateAccessToken(userID, secret)
	if err != nil {
		return
	}
	refreshToken, err = utils.GenerateRefreshToken(userID, secret)
	return
}
