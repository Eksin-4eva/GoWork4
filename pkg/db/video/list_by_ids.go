package video

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// ListByIDs 按主键批量查询视频，用于给热门榜的 ID 列表补全内容。
//
// 返回顺序不保证与入参一致，最终顺序由调用方按排行榜的顺序重排。
func (d *DBVideo) ListByIDs(ctx context.Context, ids []int64) ([]*model.Video, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	items := make([]*model.Video, 0, len(ids))
	err := d.client.WithContext(ctx).Table(videoTable).Where("id IN ?", ids).Find(&items).Error
	if err != nil {
		logger.Errorf("dal.video.ListByIDs: %v", err)
		return nil, errno.InternalError
	}
	return items, nil
}
