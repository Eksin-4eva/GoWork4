package chat

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Save 落库一条聊天消息。
func (d *DBMessage) Save(ctx context.Context, m *model.ChatMessage) error {
	err := d.client.WithContext(ctx).Table(messageTable).Create(m).Error
	if err != nil {
		logger.Errorf("dal.chat.Save: %v", err)
		return errno.InternalError
	}
	return nil
}
