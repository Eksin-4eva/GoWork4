package chat

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ListConversations 列出某个用户参与过的会话，每条会话取最新一条消息。
//
// 会话没有独立表：room_id 由双方用户 ID 排序拼接而成，
// 这里用「每个 room 的最大雪花 ID」定位最新消息，再分页返回。
func (d *DBMessage) ListConversations(
	ctx context.Context, userID int64, pageNum, pageSize int,
) ([]*model.ChatMessage, int64, error) {
	pageNum, pageSize = utils.NormalizePage(pageNum, pageSize)

	var total int64
	countSQL := `
SELECT COUNT(DISTINCT room_id)
FROM ` + messageTable + `
WHERE sender_id = ? OR receiver_id = ?`
	if err := d.client.WithContext(ctx).Raw(countSQL, userID, userID).Scan(&total).Error; err != nil {
		logger.Errorf("dal.chat.ListConversations count: %v", err)
		return nil, 0, errno.InternalError
	}

	items := make([]*model.ChatMessage, 0, pageSize)
	listSQL := `
SELECT m.id, m.room_id, m.sender_id, m.receiver_id, m.message_type, m.content,
       m.read_at, m.created_at, m.updated_at, m.deleted_at
FROM ` + messageTable + ` AS m
INNER JOIN (
	SELECT room_id, MAX(id) AS max_id
	FROM ` + messageTable + `
	WHERE sender_id = ? OR receiver_id = ?
	GROUP BY room_id
) AS latest ON m.room_id = latest.room_id AND m.id = latest.max_id
ORDER BY m.created_at DESC, m.id DESC
LIMIT ? OFFSET ?`
	err := d.client.WithContext(ctx).Raw(
		listSQL, userID, userID, pageSize, utils.Offset(pageNum, pageSize),
	).Scan(&items).Error
	if err != nil {
		logger.Errorf("dal.chat.ListConversations: %v", err)
		return nil, 0, errno.InternalError
	}
	return items, total, nil
}
