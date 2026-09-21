package comment

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ListByVideo 分页查询某个视频（或某条评论下）的评论，返回列表与总数。
//
// parentID 传 0 表示查顶层评论，非 0 表示查该评论的子评论。
func (d *DBComment) ListByVideo(
	ctx context.Context, videoID, parentID int64, pageNum, pageSize int,
) ([]*model.Comment, int64, error) {
	pageNum, pageSize = utils.NormalizePage(pageNum, pageSize)

	var total int64
	countTx := d.client.WithContext(ctx).Table(commentTable).
		Where(commentTable+".video_id = ? AND "+commentTable+".parent_id = ?", videoID, parentID)
	if err := countTx.Count(&total).Error; err != nil {
		logger.Errorf("dal.comment.ListByVideo count: %v", err)
		return nil, 0, errno.InternalError
	}

	items := make([]*model.Comment, 0, pageSize)
	err := d.client.WithContext(ctx).Table(commentTable).
		Where(commentTable+".video_id = ? AND "+commentTable+".parent_id = ?", videoID, parentID).
		Order(orderLatest).
		Limit(pageSize).
		Offset(utils.Offset(pageNum, pageSize)).
		Find(&items).Error
	if err != nil {
		logger.Errorf("dal.comment.ListByVideo: %v", err)
		return nil, 0, errno.InternalError
	}
	return items, total, nil
}
