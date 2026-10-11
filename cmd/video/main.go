// cmd/video 启动视频微服务。
//
// 它比用户服务多两个依赖：Redis（热门榜）与 Kafka（封面事件）。
package main

import (
	"context"
	"flag"
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"

	"gobili/config"
	"gobili/internal/video"
	"gobili/kitex_gen/video/videoservice"
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
	logger.Init(cfg.Services.Video.Name, cfg.Server.LogLevel)

	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID,
		base.WithMySQL(), base.WithRedis(), base.WithKafka())
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}
	defer clientset.Close()

	// 视频服务要发封面事件，所以由它确保 topic 存在
	ensureCtx, cancelEnsure := context.WithTimeout(context.Background(), kafkaSetupTimeout)
	if err := clientset.MQ.EnsureTopics(
		ensureCtx, constants.KafkaTopics(), constants.TopicPartitions); err != nil {
		logger.Warnf("ensure kafka topics: %v", err)
	}
	cancelEnsure()

	registry, err := etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		logger.Fatalf("init etcd registry failed: %v", err)
	}

	addr := &net.TCPAddr{
		IP:   net.ParseIP(cfg.Services.Video.Host),
		Port: cfg.Services.Video.Port,
	}
	svr := videoservice.NewServer(
		video.NewVideoServiceImpl(),
		server.WithServiceAddr(addr),
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: cfg.Services.Video.Name,
		}),
	)

	logger.Infof("%s listening on %s:%d", cfg.Services.Video.Name,
		cfg.Services.Video.Host, cfg.Services.Video.Port)
	if err := svr.Run(); err != nil {
		logger.Fatalf("video service run failed: %v", err)
	}
}
