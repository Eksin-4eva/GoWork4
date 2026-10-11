package video

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ListPopular 从数据库按播放数分页取热门视频。
//
// 这是 Redis 排行榜缓存不可用时的回退路径。
func (d *DBVideo) ListPopular(ctx context.Context, pageNum, pageSize int) ([]*model.Video, error) {
	pageNum, pageSize = utils.NormalizePage(pageNum, pageSize)
	return d.ListTopByVisit(ctx, pageSize, utils.Offset(pageNum, pageSize))
}

// ListTopByVisit 按播放数倒序取 TOP N，不受分页上限约束。
//
// 供热门榜缓存重建使用：重建一次要覆盖几百条，
// 走 ListPopular 会被 MaxPageSize 夹到 50 条，榜单永远只有 50 个视频。
func (d *DBVideo) ListTopByVisit(ctx context.Context, limit, offset int) ([]*model.Video, error) {
	if limit <= 0 {
		return nil, nil
	}

	items := make([]*model.Video, 0, limit)
	err := d.client.WithContext(ctx).Table(videoTable).
		Order(videoTable + ".visit_count DESC").
		Order(orderLatest).
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	if err != nil {
		logger.Errorf("dal.video.ListTopByVisit: %v", err)
		return nil, errno.InternalError
	}
	return items, nil
}
