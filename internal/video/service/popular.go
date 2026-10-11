package service

import (
	"context"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// Popular 返回热门视频。
//
// 三级策略：
//  1. 榜单已初始化 → 读 Redis
//  2. 榜单尚未初始化 → 从数据库重建后再读 Redis
//  3. Redis 不可用 → 直接查数据库
//
// 第 3 条是关键：缓存故障应当只影响性能，不该影响可用性。
func (s *VideoService) Popular(ctx context.Context, pageNum, pageSize int) ([]*model.Video, int64, error) {
	pageNum, pageSize = utils.NormalizePage(pageNum, pageSize)

	if !s.popularReady(ctx) {
		s.rebuildPopular(ctx)
	}

	if videos, total, ok := s.popularFromCache(ctx, pageNum, pageSize); ok {
		return videos, total, nil
	}

	videos, err := s.db.Video.ListPopular(ctx, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.db.Video.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return videos, total, nil
}

// popularReady 判断榜单是否可读。Redis 故障时返回 false，调用方会走数据库。
func (s *VideoService) popularReady(ctx context.Context) bool {
	ready, err := s.cache.Popular.IsReady(ctx)
	if err != nil {
		logger.Warnf("video.Popular: 读取榜单状态失败，降级查库: %v", err)
		return false
	}
	return ready
}

// popularFromCache 尝试从榜单读出这一页，任一环节失败都返回 ok=false 交给上层降级。
func (s *VideoService) popularFromCache(
	ctx context.Context, pageNum, pageSize int,
) ([]*model.Video, int64, bool) {
	ids, err := s.cache.Popular.TopIDs(ctx, int64(utils.Offset(pageNum, pageSize)), int64(pageSize))
	if err != nil {
		logger.Warnf("video.Popular: 读取榜单失败，降级查库: %v", err)
		return nil, 0, false
	}
	if len(ids) == 0 {
		// 榜单已初始化但这一页没数据：可能是翻过头了，也可能确实一条视频都没有。
		// 两种情况都该返回空列表，而不是回退查库。
		return []*model.Video{}, 0, true
	}

	videos, err := s.db.Video.ListByIDs(ctx, ids)
	if err != nil {
		return nil, 0, false
	}

	total, err := s.cache.Popular.Count(ctx)
	if err != nil {
		logger.Warnf("video.Popular: 读取榜单总数失败: %v", err)
		total = int64(len(videos))
	}
	return reorderByIDs(videos, ids), total, true
}

// rebuildPopular 用数据库里的播放量重建榜单。失败只记日志，不影响本次请求。
func (s *VideoService) rebuildPopular(ctx context.Context) {
	items, err := s.db.Video.ListTopByVisit(ctx, constants.PopularRankSize, 0)
	if err != nil {
		logger.Warnf("video.Popular: 重建榜单时读取数据库失败: %v", err)
		return
	}

	scores := make(map[int64]int64, len(items))
	for _, v := range items {
		scores[v.ID] = v.VisitCount
	}
	if err := s.cache.Popular.Replace(ctx, scores); err != nil {
		logger.Warnf("video.Popular: 写入榜单失败: %v", err)
	}
}

// reorderByIDs 把数据库结果按榜单顺序重排。
//
// 数据库的 IN 查询不保证返回顺序，不重排的话排行榜就是乱的。
func reorderByIDs(videos []*model.Video, ids []int64) []*model.Video {
	byID := make(map[int64]*model.Video, len(videos))
	for _, v := range videos {
		byID[v.ID] = v
	}

	ordered := make([]*model.Video, 0, len(ids))
	for _, id := range ids {
		if v, ok := byID[id]; ok {
			ordered = append(ordered, v)
		}
	}
	return ordered
}
