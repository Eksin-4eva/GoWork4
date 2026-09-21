// Package user 提供 users 与 refresh_tokens 两张表的数据访问。
//
// 约定（全项目 DAL 通用，对齐 fzuhelper）：
//   - 一个方法一个文件
//   - 表名通过 constants.XxxTableName 显式指定
//   - 查不到记录不算错误，返回 (false, nil, nil)，由业务层决定怎么表达
//   - 原始错误只写日志，返回给上层的一律是干净的 errno（否则会把 SQL 细节透给客户端）
package user

import (
	"gorm.io/gorm"

	"gobili/pkg/utils"
)

// DBUser 是用户域的数据访问对象。
type DBUser struct {
	client *gorm.DB
	sf     *utils.Snowflake
}

// NewDBUser 创建用户域 DAL。
func NewDBUser(client *gorm.DB, sf *utils.Snowflake) *DBUser {
	return &DBUser{client: client, sf: sf}
}
