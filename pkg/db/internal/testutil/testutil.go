// Package testutil 提供 DAL 集成测试的公共装置。
//
// 这里的测试连**真实 MySQL**，不做 mock：DAL 的价值恰恰在 SQL 本身
// （软删除是否生效、唯一键冲突如何翻译、批量更新是否原子），
// 把数据库 mock 掉就等于什么都没测。
//
// 未设置 GOBILI_TEST_DSN 时全部跳过，保证没起数据库的机器上 make test 依然全绿。
package testutil

import (
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"

	"gobili/pkg/utils"
)

// EnvDSN 是集成测试使用的 MySQL 连接串环境变量名。
//
// **务必指向独立的测试库**，不要用开发库——用例会 TRUNCATE 表。
// 先执行 make test-db 建好 gobili_test，再：
//
//	make test-integration
const EnvDSN = "GOBILI_TEST_DSN"

// Open 连接测试库；未配置 DSN 时跳过当前测试。
func Open(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("未设置 %s，跳过 DAL 集成测试", EnvDSN)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 与生产配置一致：让 MySQL 错误被翻译成 gorm.ErrDuplicatedKey 等
		TranslateError: true,
		Logger:         glogger.Default.LogMode(glogger.Silent),
	})
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	return db
}

// Snowflake 返回一个固定节点的 ID 生成器。
func Snowflake(t *testing.T) *utils.Snowflake {
	t.Helper()

	sf, err := utils.NewSnowflake(0, 0)
	if err != nil {
		t.Fatalf("创建雪花生成器失败: %v", err)
	}
	return sf
}

// Clean 清空指定表，避免用例之间互相污染。
func Clean(t *testing.T, db *gorm.DB, tables ...string) {
	t.Helper()

	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE `" + table + "`").Error; err != nil {
			t.Fatalf("清空表 %s 失败: %v", table, err)
		}
	}
}
