package video

import (
	"context"
	"sort"

	"gorm.io/gorm"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// BatchIncrementVisit 批量累加播放数，供 worker 定期 flush 聚合结果。
//
// key 为视频 ID，value 为该批次累计的增量（必须为正）。
// 放在一个事务里提交：要么这批计数全部生效，要么全部重来，
// 避免 worker 中途崩溃导致部分视频计数丢失。
func (d *DBVideo) BatchIncrementVisit(ctx context.Context, deltas map[int64]int64) error {
	if len(deltas) == 0 {
		return nil
	}

	// 排序后再更新：map 遍历顺序随机，固定顺序能避免多 worker 并发时的加锁顺序死锁
	ids := make([]int64, 0, len(deltas))
	for id := range deltas {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	err := d.client.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			delta := deltas[id]
			if delta <= 0 {
				continue
			}
			if err := tx.Table(videoTable).
				Where("id = ? AND "+notDeleted, id).
				UpdateColumn("visit_count", gorm.Expr("visit_count + ?", delta)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Errorf("dal.video.BatchIncrementVisit: %v", err)
		return errno.InternalError
	}
	return nil
}
