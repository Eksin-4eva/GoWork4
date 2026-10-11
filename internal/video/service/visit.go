package service

import (
	"context"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Visit 记录一次播放。
//
// 数据库是真源、Redis 榜单是缓存：先写库，成功后再更新榜单。
// 榜单更新失败只记日志——为了一个排行分数让播放请求失败是不划算的。
func (s *VideoService) Visit(ctx context.Context, videoID int64) error {
	exists, err := s.db.Video.IncrementVisit(ctx, videoID)
	if err != nil {
		return err
	}
	if !exists {
		return errno.BizNotExist.WithMessage("视频不存在")
	}

	if err := s.cache.Popular.Incr(ctx, videoID, 1); err != nil {
		logger.Warnf("video.Visit: 更新热门榜失败: %v", err)
	}
	return nil
}
