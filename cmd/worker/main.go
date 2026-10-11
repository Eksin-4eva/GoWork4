// cmd/worker 启动 Kafka 消费者进程。
//
// 与 api 分开部署：抽帧是 CPU 密集操作，放在网关进程里会跟接口抢资源。
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gobili/config"
	"gobili/internal/worker"
	"gobili/pkg/base"
	"gobili/pkg/constants"
	"gobili/pkg/logger"
)

// kafkaSetupTimeout 是启动时确保 topic 存在的超时。
const kafkaSetupTimeout = 5 * time.Second

var configFile = flag.String("f", "config/config.yaml", "the config file path")

func main() {
	flag.Parse()

	if err := config.Init(*configFile); err != nil {
		panic(err)
	}
	cfg := config.Get()
	logger.Init(constants.ServiceWorker, cfg.Server.LogLevel)

	// worker 需要全部依赖：抽帧要对象存储、计数要数据库与 Redis、
	// 消费要 Kafka
	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID,
		base.WithMySQL(), base.WithRedis(), base.WithMinIO(), base.WithKafka())
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}
	defer clientset.Close()

	// 消费端也确保 topic 存在：worker 先于 api 启动时，
	// 不能等 api 来建 topic，否则这段时间的事件全丢
	ensureCtx, cancelEnsure := context.WithTimeout(context.Background(), kafkaSetupTimeout)
	if err := clientset.MQ.EnsureTopics(
		ensureCtx, constants.KafkaTopics(), constants.TopicPartitions); err != nil {
		logger.Warnf("ensure kafka topics: %v", err)
	}
	cancelEnsure()

	ctx, cancel := context.WithCancel(context.Background())
	w := worker.New()

	var wg sync.WaitGroup
	for _, run := range []func(context.Context){
		w.RunCoverGenerator,
		w.RunCoverCompensator,
		w.RunVisitAggregator,
	} {
		wg.Add(1)
		go func(run func(context.Context)) {
			defer wg.Done()
			run(ctx)
		}(run)
	}

	logger.Infof("%s started", constants.ServiceWorker)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 先取消让消费协程退出，它们各自负责把缓冲区 flush 掉，
	// 所以这里必须等 wg 结束，不能直接退出进程。
	logger.Infof("%s shutting down", constants.ServiceWorker)
	cancel()
	wg.Wait()
}
