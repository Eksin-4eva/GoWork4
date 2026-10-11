// cmd/comment 启动评论微服务。它只需要数据库。
package main

import (
	"flag"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"

	"gobili/config"
	"gobili/internal/comment"
	"gobili/kitex_gen/comment/commentservice"
	"gobili/pkg/base"
	"gobili/pkg/logger"
)

var configFile = flag.String("f", "config/config.yaml", "the config file path")

func main() {
	flag.Parse()

	if err := config.Init(*configFile); err != nil {
		panic(err)
	}
	cfg := config.Get()
	logger.Init(cfg.Services.Comment.Name, cfg.Server.LogLevel)

	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID,
		base.WithMySQL())
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}
	defer clientset.Close()

	registry, err := etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		logger.Fatalf("init etcd registry failed: %v", err)
	}

	addr := &net.TCPAddr{
		IP:   net.ParseIP(cfg.Services.Comment.Host),
		Port: cfg.Services.Comment.Port,
	}
	svr := commentservice.NewServer(
		comment.NewCommentServiceImpl(),
		server.WithServiceAddr(addr),
		server.WithRegistry(registry),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: cfg.Services.Comment.Name,
		}),
	)

	logger.Infof("%s listening on %s:%d", cfg.Services.Comment.Name,
		cfg.Services.Comment.Host, cfg.Services.Comment.Port)
	if err := svr.Run(); err != nil {
		logger.Fatalf("comment service run failed: %v", err)
	}
}
