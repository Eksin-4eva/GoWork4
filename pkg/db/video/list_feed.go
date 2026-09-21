package video

import (
	"context"
	"time"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// ListFeed 按创建时间倒序取最新的视频流。
//
// from 为 nil 表示不限起始时间；不为 nil 时只返回该时刻之后的视频。
func (d *DBVideo) ListFeed(ctx context.Context, from *time.Time, limit int) ([]*model.Video, error) {
	tx := d.client.WithContext(ctx).Table(videoTable)
	if from != nil {
		tx = tx.Where(videoTable+".created_at >= ?", *from)
	}

	items := make([]*model.Video, 0, limit)
	if err := tx.Order(orderLatest).Limit(limit).Find(&items).Error; err != nil {
		logger.Errorf("dal.video.ListFeed: %v", err)
		return nil, errno.InternalError
	}
	return items, nil
}
