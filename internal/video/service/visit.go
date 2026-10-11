package service

import (
	"context"
	"strconv"

	"gobili/internal/event"
	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Visit 记录一次播放。
//
// 计数改成异步聚合：同步写库时每次播放都要更新同一行，是全站最热的写路径。
// 现在只发一个事件，由 worker 攒批后一次提交。
//
// 但存在性校验不能让：以前是靠 UPDATE 的 RowsAffected 判断视频是否存在，
// 改异步后这个副作用就没了，所以这里单独查一次（走主键索引，代价很小）。
// 不查的话，任意 video_id 都能把计数写进系统。
func (s *VideoService) Visit(ctx context.Context, videoID int64) error {
	found, _, err := s.db.Video.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if !found {
		return errno.BizNotExist.WithMessage("视频不存在")
	}

	eventID, err := s.sf.NextVal()
	if err != nil {
		logger.Errorf("video.Visit: snowflake: %v", err)
		return errno.InternalError
	}

	payload := event.VideoVisited{
		V:       constants.SchemaVersion,
		EventID: strconv.FormatInt(eventID, 10),
		VideoID: videoID,
	}
	if err := s.mq.Producer.Send(ctx, constants.TopicVideoVisited,
		[]byte(strconv.FormatInt(videoID, 10)), payload); err != nil {
		// 发不出去就同步补一次：宁可退化回原来的写热点，也不能静默丢计数
		logger.Warnf("video.Visit: 发送播放事件失败，降级为同步写: %v", err)
		return s.recordVisitSync(ctx, videoID)
	}
	return nil
}

// recordVisitSync 是消息队列不可用时的降级路径。
func (s *VideoService) recordVisitSync(ctx context.Context, videoID int64) error {
	if _, err := s.db.Video.IncrementVisit(ctx, videoID); err != nil {
		return err
	}
	if err := s.cache.Popular.Incr(ctx, videoID, 1); err != nil {
		logger.Warnf("video.Visit: 更新热门榜失败: %v", err)
	}
	return nil
}
