package video

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// ListMissingCover 找出还没有封面的视频，供补偿任务重发事件。
//
// 只扫描 cover_key 为空的行，走 idx_cover_key 索引，不受表大小影响。
func (d *DBVideo) ListMissingCover(ctx context.Context, limit int) ([]*model.Video, error) {
	items := make([]*model.Video, 0, limit)
	err := d.client.WithContext(ctx).Table(videoTable).
		Where(videoTable + ".cover_key = ''").
		Order(videoTable + ".created_at ASC").
		Limit(limit).
		Find(&items).Error
	if err != nil {
		logger.Errorf("dal.video.ListMissingCover: %v", err)
		return nil, errno.InternalError
	}
	return items, nil
}
