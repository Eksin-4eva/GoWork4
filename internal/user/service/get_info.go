package service

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
)

// GetInfo 查询用户信息。targetUserID 为 0 表示查当前登录用户。
func (s *UserService) GetInfo(ctx context.Context, currentUserID, targetUserID int64) (*model.User, error) {
	if targetUserID == 0 {
		targetUserID = currentUserID
	}
	if targetUserID == 0 {
		return nil, errno.AuthInvalid
	}

	found, u, err := s.db.User.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errno.BizNotExist.WithMessage("用户不存在")
	}
	return u, nil
}
