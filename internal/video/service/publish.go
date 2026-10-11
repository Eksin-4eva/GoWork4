package service

import (
	"context"
	"strings"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// PublishParams 是投稿入参。
type PublishParams struct {
	Title       string
	Description string
	Video       utils.Upload
	// Cover 可选。不传时 cover_key 留空，由 Phase 8 的 worker 异步抽帧补上。
	Cover *utils.Upload
}

// Publish 投稿视频。
//
// 文件的 object key 里带视频 ID，所以必须先生成 ID 再上传。
func (s *VideoService) Publish(ctx context.Context, userID int64, p PublishParams) error {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return errno.ParamEmpty.WithMessage("标题不能为空")
	}
	if len([]rune(title)) > constants.MaxVideoTitleRunes {
		return errno.ParamError.WithMessage("标题不能超过 128 个字")
	}

	id, err := s.sf.NextVal()
	if err != nil {
		logger.Errorf("video.Publish: snowflake: %v", err)
		return errno.InternalError
	}

	videoKey, err := s.storage.Put(ctx, constants.StoragePrefixVideo, id, utils.AllowedVideoMIME, p.Video, "视频")
	if err != nil {
		return err
	}

	coverKey := ""
	if p.Cover != nil {
		coverKey, err = s.storage.Put(ctx, constants.StoragePrefixCover, id, utils.AllowedImageMIME, *p.Cover, "封面")
		if err != nil {
			return err
		}
	}

	return s.db.Video.Create(ctx, &model.Video{
		ID:          id,
		UserID:      userID,
		VideoKey:    videoKey,
		CoverKey:    coverKey,
		Title:       title,
		Description: strings.TrimSpace(p.Description),
	})
}
