package video

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ListByUser 分页查询某个用户的投稿，返回列表与总数。
func (d *DBVideo) ListByUser(
	ctx context.Context, userID int64, pageNum, pageSize int,
) ([]*model.Video, int64, error) {
	pageNum, pageSize = utils.NormalizePage(pageNum, pageSize)

	var total int64
	countTx := d.client.WithContext(ctx).Table(videoTable).Where(videoTable+".user_id = ?", userID)
	if err := countTx.Count(&total).Error; err != nil {
		logger.Errorf("dal.video.ListByUser count: %v", err)
		return nil, 0, errno.InternalError
	}

	items := make([]*model.Video, 0, pageSize)
	err := d.client.WithContext(ctx).Table(videoTable).
		Where(videoTable+".user_id = ?", userID).
		Order(orderLatest).
		Limit(pageSize).
		Offset(utils.Offset(pageNum, pageSize)).
		Find(&items).Error
	if err != nil {
		logger.Errorf("dal.video.ListByUser: %v", err)
		return nil, 0, errno.InternalError
	}
	return items, total, nil
}
