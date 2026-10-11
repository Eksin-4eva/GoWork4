package service

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/utils"
)

// UploadAvatar 上传头像并更新用户的 avatar_key。
//
// 顺序是先传对象存储、再写数据库：写库失败最多在存储里留一个孤儿对象；
// 反过来则会留下「库里有 key、存储里没有文件」的坏数据，前端直接裂图。
func (s *UserService) UploadAvatar(
	ctx context.Context, userID int64, file utils.Upload,
) (*model.User, error) {
	key, err := s.storage.Put(ctx, constants.StoragePrefixAvatar, userID,
		utils.AllowedImageMIME, file, "头像")
	if err != nil {
		return nil, err
	}

	if err := s.db.User.UpdateAvatarKey(ctx, userID, key); err != nil {
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
