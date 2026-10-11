// Package base 聚合服务运行期所需的全部外部依赖。
package base

import (
	"fmt"

	"gobili/config"
	"gobili/pkg/base/client"
	"gobili/pkg/cache"
	"gobili/pkg/db"
	"gobili/pkg/kafka"
	"gobili/pkg/logger"
	"gobili/pkg/storage"
	"gobili/pkg/utils"
)

// instance 是进程级唯一的依赖集合。
//
// 业务层拿不到 handler 的参数，所以只能通过包级单例访问依赖，
// 由 main 在启动时初始化一次。
var instance *ClientSet

// ClientSet 是可注入给业务层的依赖集合。
//
// 拆成微服务之后，不同进程需要的依赖完全不同：
//
//	api      只有对象存储（文件从这里上传，RPC 只传 key）
//	user     只有数据库
//	video    数据库 + Redis + Kafka
//	comment  只有数据库
//	worker   全部
//	chat     只有数据库
//
// 所以用 Option 按需初始化，而不是一次性全建。
// **否则「用户服务」会因为「Redis 没起」而启动失败**——它根本不需要 Redis。
type ClientSet struct {
	DB      *db.Database
	Cache   *cache.Cache
	Storage *storage.Storage
	SF      *utils.Snowflake
	MQ      *kafka.Broker

	cleanups []func()
}

// Option 初始化某一个依赖。
type Option func(cs *ClientSet) error

// NewClientSet 按传入的 Option 初始化依赖。
//
// 雪花生成器没有外部依赖，所以总是创建——几乎所有业务都要用它生成 ID。
func NewClientSet(datacenterID, workerID int64, opts ...Option) (*ClientSet, error) {
	sf, err := utils.NewSnowflake(datacenterID, workerID)
	if err != nil {
		return nil, fmt.Errorf("base.NewClientSet: new snowflake: %w", err)
	}

	cs := &ClientSet{SF: sf}
	for _, opt := range opts {
		if err := opt(cs); err != nil {
			return nil, err
		}
	}

	instance = cs
	return cs, nil
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

// WithMySQL 初始化数据库连接与各业务域 DAL。
func WithMySQL() Option {
	return func(cs *ClientSet) error {
		gormDB, err := client.InitMySQL()
		if err != nil {
			return err
		}
		logger.Infof("clientset: mysql connected")

		cs.DB = db.NewDatabase(gormDB, cs.SF)
		cs.cleanups = append(cs.cleanups, func() {
			sqlDB, err := gormDB.DB()
			if err != nil {
				logger.Errorf("clientset: get sql.DB: %v", err)
				return
			}
			if err := sqlDB.Close(); err != nil {
				logger.Errorf("clientset: close mysql: %v", err)
			}
		})
		return nil
	}
}

// WithRedis 初始化缓存（热榜、事件去重）。
func WithRedis() Option {
	return func(cs *ClientSet) error {
		rdb, err := client.InitRedis()
		if err != nil {
			return err
		}
		logger.Infof("clientset: redis connected")

		cs.Cache = cache.New(rdb)
		cs.cleanups = append(cs.cleanups, func() {
			if err := rdb.Close(); err != nil {
				logger.Errorf("clientset: close redis: %v", err)
			}
		})
		return nil
	}
}

// WithMinIO 初始化对象存储。
func WithMinIO() Option {
	return func(cs *ClientSet) error {
		minioClient, err := client.InitMinIO()
		if err != nil {
			return err
		}
		logger.Infof("clientset: minio connected")
		cs.Storage = storage.New(minioClient)
		return nil
	}
}

// WithKafka 初始化事件的生产与消费入口。
//
// 此时不会建立连接——kafka-go 首次读写时才拨号，
// 所以 Kafka 没起也不影响服务启动。
func WithKafka() Option {
	return func(cs *ClientSet) error {
		cfg := config.Get()
		cs.MQ = kafka.NewBroker(cfg.Kafka.Brokers, cfg.Kafka.GroupID)
		logger.Infof("clientset: kafka broker configured (%v)", cfg.Kafka.Brokers)
		return nil
	}
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
