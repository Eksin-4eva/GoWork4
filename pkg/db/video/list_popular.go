package video

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ListPopular 从数据库按播放数取热门视频。
//
// 这是 Redis 排行榜缓存未命中时的回退路径，正常情况走 cache.Popular。
func (d *DBVideo) ListPopular(ctx context.Context, pageNum, pageSize int) ([]*model.Video, error) {
	pageNum, pageSize = utils.NormalizePage(pageNum, pageSize)

	items := make([]*model.Video, 0, pageSize)
	err := d.client.WithContext(ctx).Table(videoTable).
		Order(videoTable + ".visit_count DESC").
		Order(orderLatest).
		Limit(pageSize).
		Offset(utils.Offset(pageNum, pageSize)).
		Find(&items).Error
	if err != nil {
		logger.Errorf("dal.video.ListPopular: %v", err)
		return nil, errno.InternalError
	}
	return items, nil
}
