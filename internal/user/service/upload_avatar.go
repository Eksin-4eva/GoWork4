package service

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// UploadAvatar 上传头像并更新用户的 avatar_key。
//
// 顺序是先传对象存储、再写数据库：写库失败最多在存储里留一个孤儿对象；
// 反过来则会留下「库里有 key、存储里没有文件」的坏数据，前端直接裂图。
func (s *UserService) UploadAvatar(
	ctx context.Context, userID int64, file utils.Upload,
) (*model.User, error) {
	if file.Size <= 0 {
		return nil, errno.ParamFile.WithMessage("文件为空")
	}
	if file.Size > constants.MaxUploadMB<<20 {
		return nil, errno.ParamFile.WithMessage("头像不能超过 100MB")
	}

	sniffed, err := utils.DetectContentType(file.Reader)
	if err != nil {
		logger.Errorf("user.UploadAvatar: detect content type: %v", err)
		return nil, errno.ParamFile.WithMessage("无法识别文件类型")
	}
	ext, err := utils.PickExtension(sniffed, file.Filename, utils.AllowedImageMIME)
	if err != nil {
		return nil, errno.ParamFile.WithError(err)
	}

	// key 只跟用户绑定，重复上传直接覆盖，不会在存储里堆垃圾
	key := fmt.Sprintf("%s/%d%s", constants.StoragePrefixAvatar, userID, ext)
	_, err = s.storage.PutObject(ctx, constants.StorageBucket, key, file.Reader, file.Size,
		minio.PutObjectOptions{ContentType: sniffed})
	if err != nil {
		logger.Errorf("user.UploadAvatar: put object: %v", err)
		return nil, errno.InternalError
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
