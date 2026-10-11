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
)

// PublishVideoParams 是投稿入参。
//
// 视频与封面已经是 object key（由网关上传递到对象存储后传进来），
// 本服务不接触文件本身。
type PublishVideoParams struct {
	Title       string
	Description string
	VideoKey    string
	// CoverKey 为空表示由 worker 异步抽帧生成
	CoverKey string
}

// Publish 投稿视频。
func (s *VideoService) Publish(ctx context.Context, userID int64, p PublishVideoParams) error {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return errno.ParamEmpty.WithMessage("标题不能为空")
	}
	if len([]rune(title)) > constants.MaxVideoTitleRunes {
		return errno.ParamError.WithMessage("标题不能超过 128 个字")
	}
	if p.VideoKey == "" {
		return errno.ParamFile.WithMessage("缺少视频文件")
	}

	id, err := s.sf.NextVal()
	if err != nil {
		logger.Errorf("video.Publish: snowflake: %v", err)
		return errno.InternalError
	}

	video := &model.Video{
		ID:          id,
		UserID:      userID,
		VideoKey:    p.VideoKey,
		CoverKey:    p.CoverKey,
		Title:       title,
		Description: strings.TrimSpace(p.Description),
	}
	if err := s.db.Video.Create(ctx, video); err != nil {
		return err
	}

	// 客户端没传封面时交给 worker 异步抽帧；传了就同步存好了，不必再排一次队
	if video.CoverKey == "" {
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
