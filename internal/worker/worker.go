// Package worker 消费 Kafka 事件并完成异步副作用。
//
// 与 api 分开部署：抽帧是 CPU 密集操作，放在网关进程里会跟接口抢资源；
// worker 崩了也不影响用户请求，重启后从上次的 offset 继续。
package worker

import (
	"gobili/config"
	"gobili/pkg/base"
	"gobili/pkg/cache"
	"gobili/pkg/db"
	"gobili/pkg/kafka"
	"gobili/pkg/storage"
)

// Worker 持有消费端需要的全部依赖。
type Worker struct {
	db        *db.Database
	cache     *cache.Cache
	storage   *storage.Storage
	mq        *kafka.Broker
	extractor Extractor
	visits    *visitAggregator
}

// New 构造 worker。依赖取自进程级 ClientSet。
func New() *Worker {
	cs := base.GetClientSet()
	return newWorker(cs.DB, cs.Cache, cs.Storage, cs.MQ,
		FFmpeg{Binary: config.Get().Kafka.FfmpegPath})
}

// newWorker 用给定依赖构造 worker，供测试替换掉真实的 ffmpeg。
func newWorker(
	database *db.Database, c *cache.Cache, st *storage.Storage,
	mq *kafka.Broker, extractor Extractor,
) *Worker {
	return &Worker{
		db:        database,
		cache:     c,
		storage:   st,
		mq:        mq,
		extractor: extractor,
		visits:    newVisitAggregator(database, c),
	}
}
