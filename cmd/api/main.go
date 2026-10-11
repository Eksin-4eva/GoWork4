// cmd/api 启动 HTTP 网关。
//
// 微服务化之后网关的职责变得很纯粹：
//
//	解析 HTTP → 必要时把文件传到对象存储 → 调用下游 RPC → 拼响应
//
// 它**不再直连数据库、Redis、Kafka**。这不只是「少几个连接」，
// 而是意味着数据访问的边界清晰了：所有 SQL 都在微服务里，
// 网关想绕过服务直接改数据都做不到。
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	"gobili/api/router"
	"gobili/api/rpc"
	"gobili/config"
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
	logger.Init(cfg.Server.Name, cfg.Server.LogLevel)

	// 网关只需要对象存储：上传的文件在这里落 MinIO，
	// 下游服务只接收 object key
	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID,
		base.WithMinIO())
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}

	// 初始化 RPC 客户端：地址从 etcd 动态发现，不写死在配置里
	rpc.Init()

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
