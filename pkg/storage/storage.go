// Package storage 在 MinIO 客户端之上提供项目级的对象存取封装。
//
// 上传前的校验（大小上限、真实类型嗅探、扩展名白名单）与 object key 的生成规则
// 只在这里实现一次，用户头像与视频投稿共用同一套，避免各写一份导致规则漂移。
//
// 注意这里只负责**写**；key 反向拼成对外 URL 是展示层的事，在 api/pack 里。
package storage

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// Storage 是对象存储的读写入口。
type Storage struct {
	client *minio.Client
}

// New 构造封装。
func New(client *minio.Client) *Storage {
	return &Storage{client: client}
}

// Download 把对象下载到本地文件。worker 抽帧前需要先把视频落到磁盘。
func (s *Storage) Download(ctx context.Context, key, destPath string) error {
	if err := s.client.FGetObject(ctx, constants.StorageBucket, key, destPath,
		minio.GetObjectOptions{}); err != nil {
		return fmt.Errorf("storage.Download %s: %w", key, err)
	}
	return nil
}

// Put 校验并上传一个文件，返回 object key。
//
// key 形如 <prefix>/<ownerID><ext>：前缀区分资源类型，ownerID 让同一主体的
// 重复上传直接覆盖，不会在存储里堆垃圾。
//
// label 只用于错误文案（"头像" / "视频" / "封面"）。
func (s *Storage) Put(
	ctx context.Context, prefix string, ownerID int64,
	allowed map[string]string, file utils.Upload, label string,
) (string, error) {
	if file.Size <= 0 {
		return "", errno.ParamFile.WithMessage(label + "文件为空")
	}
	if file.Size > constants.MaxUploadMB<<20 {
		return "", errno.ParamFile.WithMessage(
			fmt.Sprintf("%s不能超过 %dMB", label, constants.MaxUploadMB))
	}

	sniffed, err := utils.DetectContentType(file.Reader)
	if err != nil {
		logger.Errorf("storage.Put: detect content type: %v", err)
		return "", errno.ParamFile.WithMessage("无法识别" + label + "类型")
	}
	ext, err := utils.PickExtension(sniffed, file.Filename, allowed)
	if err != nil {
		return "", errno.ParamFile.WithError(err)
	}

	key := fmt.Sprintf("%s/%d%s", prefix, ownerID, ext)
	_, err = s.client.PutObject(ctx, constants.StorageBucket, key, file.Reader, file.Size,
		minio.PutObjectOptions{ContentType: sniffed})
	if err != nil {
		logger.Errorf("storage.Put: put object %s: %v", key, err)
		return "", errno.InternalError
	}
	return key, nil
}
