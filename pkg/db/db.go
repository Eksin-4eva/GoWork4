// Package db 聚合各业务域的 DAL。
//
// 业务层只依赖 *db.Database，不直接持有 *gorm.DB——
// 这样 SQL 全部收敛在 pkg/db 下，将来换 ORM 也只需要改这一层。
package db

import (
	"gorm.io/gorm"

	"gobili/pkg/db/comment"
	"gobili/pkg/db/user"
	"gobili/pkg/db/video"
	"gobili/pkg/utils"
)

// Database 是所有数据访问对象的入口。
type Database struct {
	User    *user.DBUser
	Video   *video.DBVideo
	Comment *comment.DBComment
}

// NewDatabase 构造各业务域的 DAL。
func NewDatabase(client *gorm.DB, sf *utils.Snowflake) *Database {
	return &Database{
		User:    user.NewDBUser(client, sf),
		Video:   video.NewDBVideo(client, sf),
		Comment: comment.NewDBComment(client, sf),
	}
}
