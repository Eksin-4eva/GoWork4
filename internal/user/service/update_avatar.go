package service

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
)

// UpdateAvatar 更新用户头像的 object key。
//
// 调用方（网关）已经把文件传到对象存储并拿到 key。
// **文件不走 RPC** 是刻意的设计：走 RPC 意味着文件要在内存里
// 序列化一次、反序列化一次，视频这种几十上百 MB 的文件完全不可接受。
func (s *UserService) UpdateAvatar(ctx context.Context, userID int64, avatarKey string) (*model.User, error) {
	if avatarKey == "" {
		return nil, errno.ParamEmpty.WithMessage("avatar_key 不能为空")
	}

	if err := s.db.User.UpdateAvatarKey(ctx, userID, avatarKey); err != nil {
		return nil, err
	}

	found, u, err := s.db.User.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errno.BizNotExist.WithMessage("用户不存在")
	}
	return u, nil
}
