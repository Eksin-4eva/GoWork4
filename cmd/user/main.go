// cmd/user 启动用户微服务。
//
// 与其他微服务的差别只有两处：注册的服务名，以及它需要的依赖
// （用户服务只连数据库，不连 Redis / Kafka / 对象存储）。
package main

import (
	"flag"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"

	"gobili/config"
	"gobili/internal/user"
	"gobili/kitex_gen/user/userservice"
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
	logger.Init(cfg.Services.User.Name, cfg.Server.LogLevel)

	// 只初始化自己需要的依赖：不然 Redis 没起也会让用户服务启动失败
	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID,
		base.WithMySQL())
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}
	defer clientset.Close()

	// 把服务注册到 etcd，网关才能按服务名找到它
	registry, err := etcd.NewEtcdRegistry(cfg.Etcd.Endpoints)
	if err != nil {
		logger.Fatalf("init etcd registry failed: %v", err)
	}

	addr := &net.TCPAddr{
		IP:   net.ParseIP(cfg.Services.User.Host),
		Port: cfg.Services.User.Port,
	}
	svr := userservice.NewServer(
		user.NewUserServiceImpl(),
		server.WithServiceAddr(addr),
		server.WithRegistry(registry),
		// 注册名要与网关查找时用的名字一致
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: cfg.Services.User.Name,
		}),
	)

	logger.Infof("%s listening on %s:%d", cfg.Services.User.Name,
		cfg.Services.User.Host, cfg.Services.User.Port)
	if err := svr.Run(); err != nil {
		logger.Fatalf("user service run failed: %v", err)
	}
}
