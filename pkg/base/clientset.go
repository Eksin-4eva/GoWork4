// Package base 聚合服务运行期所需的全部外部依赖。
package base

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"

	"gobili/pkg/base/client"
	"gobili/pkg/db"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ClientSet 是可注入给业务层的依赖集合。
type ClientSet struct {
	// DB 是各业务域 DAL 的聚合入口。业务层不直接持有 *gorm.DB，
	// 所有 SQL 都收敛在 pkg/db 下。
	DB      *db.Database
	Cache   *redis.Client
	Storage *minio.Client
	SF      *utils.Snowflake

	cleanups []func()
}

// NewClientSet 初始化全部外部依赖，任一依赖不可用即返回错误（快速失败）。
func NewClientSet(datacenterID, workerID int64) (*ClientSet, error) {
	sf, err := utils.NewSnowflake(datacenterID, workerID)
	if err != nil {
		return nil, fmt.Errorf("base.NewClientSet: new snowflake: %w", err)
	}

	gormDB, err := client.InitMySQL()
	if err != nil {
		return nil, err
	}
	logger.Infof("clientset: mysql connected")

	cache, err := client.InitRedis()
	if err != nil {
		return nil, err
	}
	logger.Infof("clientset: redis connected")

	storage, err := client.InitMinIO()
	if err != nil {
		return nil, err
	}
	logger.Infof("clientset: minio connected")

	return &ClientSet{
		DB:      db.NewDatabase(gormDB, sf),
		Cache:   cache,
		Storage: storage,
		SF:      sf,
		cleanups: []func(){
			func() {
				if err := cache.Close(); err != nil {
					logger.Errorf("clientset: close redis: %v", err)
				}
			},
			func() {
				sqlDB, err := gormDB.DB()
				if err != nil {
					logger.Errorf("clientset: get sql.DB: %v", err)
					return
				}
				if err := sqlDB.Close(); err != nil {
					logger.Errorf("clientset: close mysql: %v", err)
				}
			},
		},
	}, nil
}

// Close 依次执行各资源的清理函数。清理失败只记日志，不影响调用方。
func (c *ClientSet) Close() {
	if c == nil {
		return
	}
	for _, cleanup := range c.cleanups {
		cleanup()
	}
}
