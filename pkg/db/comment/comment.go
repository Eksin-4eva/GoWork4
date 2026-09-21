// Package comment 提供 comments 表的数据访问。
package comment

import (
	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/utils"
)

// DBComment 是评论域的数据访问对象。
type DBComment struct {
	client *gorm.DB
	sf     *utils.Snowflake
}

// 表名常量。列表查询要 JOIN users 拿评论者信息时不加前缀会 ambiguous。
const (
	commentTable = constants.CommentTableName
	userTable    = constants.UserTableName

	// orderLatest 见 pkg/db/video 的同名常量：created_at 只到秒，
	// 必须用雪花 ID 兜底才能让翻页结果稳定。
	orderLatest = commentTable + ".created_at DESC, " + commentTable + ".id DESC"

	// notDeleted 必须显式写进写操作的 WHERE：
	// GORM 只给 SELECT 自动加软删条件，用 .Table() 的 UPDATE 拿不到 schema，不会自动补。
	notDeleted = commentTable + ".deleted_at IS NULL"
)

// NewDBComment 创建评论域 DAL。
func NewDBComment(client *gorm.DB, sf *utils.Snowflake) *DBComment {
	return &DBComment{client: client, sf: sf}
}
