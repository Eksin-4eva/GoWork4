// Package chat 提供 chat_messages 表的数据访问。
package chat

import (
	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/utils"
)

// DBMessage 是聊天消息的数据访问对象。
type DBMessage struct {
	client *gorm.DB
	sf     *utils.Snowflake
}

const messageTable = constants.ChatMessageTableName

// NewDBMessage 创建聊天消息 DAL。
func NewDBMessage(client *gorm.DB, sf *utils.Snowflake) *DBMessage {
	return &DBMessage{client: client, sf: sf}
}
