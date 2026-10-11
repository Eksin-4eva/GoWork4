package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"

	"gobili/api/router"
	"gobili/config"
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
	logger.Init(cfg.Server.Name, cfg.Server.LogLevel)

	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID)
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}

	// 确保 topic 存在：首次生产时 topic 不存在会直接报 UnknownTopicOrPartition。
	// 尽力而为即可，失败不该拦住启动——事件发不出去有降级与补偿兜底。
	ensureCtx, cancelEnsure := context.WithTimeout(context.Background(), kafkaSetupTimeout)
	if err := clientset.MQ.EnsureTopics(
		ensureCtx, constants.KafkaTopics(), constants.TopicPartitions); err != nil {
		logger.Warnf("ensure kafka topics: %v", err)
	}
	cancelEnsure()

	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
	)
	h.OnShutdown = append(h.OnShutdown, func(_ context.Context) {
		clientset.Close()
	})

	router.Register(h)

	logger.Infof("%s listening on %s:%d", cfg.Server.Name, cfg.Server.Host, cfg.Server.Port)
	h.Spin()
}
