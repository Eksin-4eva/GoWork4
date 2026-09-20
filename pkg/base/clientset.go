// Package base 聚合服务运行期所需的全部外部依赖。
package base

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"gobili/pkg/base/client"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// ClientSet 是可注入给业务层的依赖集合。
type ClientSet struct {
	DB      *gorm.DB
	Cache   *redis.Client
	Storage *minio.Client
	SF      *utils.Snowflake
}

// NewClientSet 初始化全部外部依赖，任一依赖不可用即返回错误（快速失败）。
func NewClientSet(datacenterID, workerID int64) (*ClientSet, error) {
	db, err := client.InitMySQL()
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

	sf, err := utils.NewSnowflake(datacenterID, workerID)
	if err != nil {
		return nil, fmt.Errorf("base.NewClientSet: new snowflake: %w", err)
	}

	return &ClientSet{DB: db, Cache: cache, Storage: storage, SF: sf}, nil
}

// Close 释放可关闭的资源。
func (c *ClientSet) Close() error {
	if c == nil {
		return nil
	}
	if c.Cache != nil {
		if err := c.Cache.Close(); err != nil {
			return fmt.Errorf("base.ClientSet.Close: close redis: %w", err)
		}
	}
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return fmt.Errorf("base.ClientSet.Close: get sql.DB: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("base.ClientSet.Close: close mysql: %w", err)
		}
	}
	return nil
}
