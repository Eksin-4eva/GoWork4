// Package video 提供 videos 表的数据访问。
package video

import (
	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/utils"
)

// DBVideo 是视频域的数据访问对象。
type DBVideo struct {
	client *gorm.DB
	sf     *utils.Snowflake
}

// 表名常量。查询一律带表名前缀，因为搜索要 JOIN users，
// 两张表都有 deleted_at，不加前缀会触发 ambiguous column 错误。
const (
	videoTable = constants.VideoTableName
	userTable  = constants.UserTableName

	// 按用户名搜索时用的 JOIN。
	// 显式补上 users.deleted_at 条件：JOIN 进来的表不会被 GORM 的软删逻辑自动覆盖。
	searchJoin = "JOIN " + userTable +
		" ON " + userTable + ".id = " + videoTable + ".user_id" +
		" AND " + userTable + ".deleted_at IS NULL"

	// orderLatest 是「最新优先」的统一排序。
	//
	// created_at 是秒精度，同一秒入库的多条记录会并列，
	// 只用它排序时结果不稳定，翻页会出现重复行或漏行。
	// 雪花 ID 高位是时间戳，用它兜底既能稳定排序又保持时间序。
	orderLatest = videoTable + ".created_at DESC, " + videoTable + ".id DESC"

	// notDeleted 是软删过滤条件。
	//
	// 注意：GORM 只会给 SELECT 自动加软删条件。UPDATE 语句因为我们用的是
	// .Table(表名) 而不是 .Model(结构体)，GORM 拿不到 schema，不会补这个条件。
	// 所以所有写操作必须自己带上它，否则软删过的行还会被改。
	notDeleted = videoTable + ".deleted_at IS NULL"
)

// NewDBVideo 创建视频域 DAL。
func NewDBVideo(client *gorm.DB, sf *utils.Snowflake) *DBVideo {
	return &DBVideo{client: client, sf: sf}
}
