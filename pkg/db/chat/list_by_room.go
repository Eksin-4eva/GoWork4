package chat

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ListByRoom 分页查询某个会话的历史消息，按时间倒序。
func (d *DBMessage) ListByRoom(
	ctx context.Context, roomID string, pageNum, pageSize int,
) ([]*model.ChatMessage, int64, error) {
	pageNum, pageSize = utils.NormalizePage(pageNum, pageSize)

	var total int64
	if err := d.client.WithContext(ctx).Table(messageTable).
		Where("room_id = ?", roomID).Count(&total).Error; err != nil {
		logger.Errorf("dal.chat.ListByRoom count: %v", err)
		return nil, 0, errno.InternalError
	}

	// created_at 只到秒，同秒消息靠雪花 ID 兜底排序，
	// 否则翻页会出现重复或漏行（与 Phase 3 的结论一致）
	items := make([]*model.ChatMessage, 0, pageSize)
	err := d.client.WithContext(ctx).Table(messageTable).
		Where("room_id = ?", roomID).
		Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset(utils.Offset(pageNum, pageSize)).
		Find(&items).Error
	if err != nil {
		logger.Errorf("dal.chat.ListByRoom: %v", err)
		return nil, 0, errno.InternalError
	}
	return items, total, nil
}
