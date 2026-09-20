// Package client 提供各类外部依赖的初始化函数。
package client

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"gobili/config"
	"gobili/pkg/constants"
	"gobili/pkg/logger"
)

// gormWriter 把 GORM 的日志转发到项目 logger。
type gormWriter struct{}

func (gormWriter) Printf(format string, args ...any) {
	logger.Debugf(strings.TrimSpace(format), args...)
}

// InitMySQL 初始化 GORM 客户端，做连通性检查并配置连接池。
func InitMySQL() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.Get().MySQL.DSN()), &gorm.Config{
		// 缓存 prepared statement，避免重复解析 SQL
		PrepareStmt: true,
		// 把 MySQL 错误翻译成 gorm.ErrDuplicatedKey 等语义化错误，
		// 业务层不必再靠匹配错误字符串判断重复键
		TranslateError: true,
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger: glogger.New(gormWriter{}, glogger.Config{
			SlowThreshold:             constants.DBSlowThreshold,
			LogLevel:                  glogger.Warn,
			IgnoreRecordNotFoundError: true,
			// 不在日志中打印参数值，避免泄露密码等敏感数据
			ParameterizedQueries: true,
			Colorful:             false,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("client.InitMySQL: open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("client.InitMySQL: get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(constants.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(constants.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(constants.DBConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("client.InitMySQL: ping mysql: %w", err)
	}
	return db, nil
}
