// Package base 聚合服务运行期所需的全部外部依赖。
package base

import (
	"fmt"

	"github.com/minio/minio-go/v7"

	"gobili/pkg/base/client"
	"gobili/pkg/cache"
	"gobili/pkg/db"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// instance 是进程级唯一的依赖集合。
//
// 业务层拿不到 handler 的参数，所以只能通过包级单例访问依赖，
// 由 main 在启动时初始化一次。这与 fzuhelper 的做法一致。
var instance *ClientSet

// ClientSet 是可注入给业务层的依赖集合。
type ClientSet struct {
	// DB 是各业务域 DAL 的聚合入口。业务层不直接持有 *gorm.DB，
	// 所有 SQL 都收敛在 pkg/db 下。
	DB *db.Database
	// Cache 是业务缓存（热榜等）。底层 Redis 客户端不直接暴露，
	// 避免业务层绕过缓存封装去裸写 key。
	Cache   *cache.Cache
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

	cacheClient, err := client.InitRedis()
	if err != nil {
		return nil, err
	}
	logger.Infof("clientset: redis connected")

	storage, err := client.InitMinIO()
	if err != nil {
		return nil, err
	}
	logger.Infof("clientset: minio connected")

	instance = &ClientSet{
		DB:      db.NewDatabase(gormDB, sf),
		Cache:   cache.New(cacheClient),
		Storage: storage,
		SF:      sf,
		cleanups: []func(){
			func() {
				if err := cacheClient.Close(); err != nil {
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
	}
	return instance, nil
}

// GetClientSet 返回已初始化的依赖集合。
//
// 未初始化时直接 panic：这是编程错误（忘了在 main 里调 NewClientSet），
// 静默返回 nil 只会把问题推迟到某个请求上。
func GetClientSet() *ClientSet {
	if instance == nil {
		panic("base.GetClientSet: clientset is not initialized, call NewClientSet in main first")
	}
	return instance
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
