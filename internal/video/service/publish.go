package service

import (
	"context"
	"strconv"
	"strings"

	"gobili/internal/event"
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

	video := &model.Video{
		ID:          id,
		UserID:      userID,
		VideoKey:    videoKey,
		CoverKey:    coverKey,
		Title:       title,
		Description: strings.TrimSpace(p.Description),
	}
	if err := s.db.Video.Create(ctx, video); err != nil {
		return err
	}

	// 客户端没传封面时交给 worker 异步抽帧；传了就同步存好了，不必再排一次队
	if coverKey == "" {
		s.emitUploaded(ctx, video)
	}
	return nil
}

// emitUploaded 发投稿事件，让 worker 异步抽帧生成封面。
//
// 发送失败只记日志：用户已经投稿成功，不能因为消息队列出问题就告诉他失败。
// 封面有补偿任务兜底（定期扫描 cover_key 为空的视频重发事件）。
func (s *VideoService) emitUploaded(ctx context.Context, v *model.Video) {
	payload := event.VideoUploaded{
		V:        constants.SchemaVersion,
		VideoID:  v.ID,
		VideoKey: v.VideoKey,
	}
	// key 用视频 ID：同一视频的事件落到同一分区，天然有序
	key := []byte(strconv.FormatInt(v.ID, 10))

	if err := s.mq.Producer.Send(ctx, constants.TopicVideoUploaded, key, payload); err != nil {
		logger.Warnf("video.Publish: 发送封面事件失败（补偿任务会重试）: %v", err)
	}
}
