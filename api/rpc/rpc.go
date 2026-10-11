// Package rpc 持有并初始化各微服务的 Kitex 客户端。
//
// 微服务化之后，网关**不再直连数据库**：它只持有 RPC 客户端与对象存储。
//
// 服务地址不写死在配置里，而是通过 etcd 动态发现——
// 微服务实例上下线、扩容，网关既不用重启也不用改配置。
// 这是服务注册中心相对「配置文件写死 IP」的核心价值。
package rpc

import (
	"sync"
	"time"

	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"

	"gobili/config"
	"gobili/kitex_gen/comment/commentservice"
	"gobili/kitex_gen/user/userservice"
	"gobili/kitex_gen/video/videoservice"
	"gobili/pkg/logger"
)

// RPC 超时。
//
// 必须有：下游服务卡住时，网关不能跟着一起挂住——
// 没有超时的调用会把网关的连接和 goroutine 慢慢耗光。
const (
	rpcTimeout     = 3 * time.Second
	connectTimeout = 1 * time.Second
)

var (
	initOnce sync.Once

	userClient    userservice.Client
	videoClient   videoservice.Client
	commentClient commentservice.Client
)

// Init 初始化全部 RPC 客户端。进程生命周期内只执行一次。
func Init() {
	initOnce.Do(func() {
		cfg := config.Get()

		resolver, err := etcd.NewEtcdResolver(cfg.Etcd.Endpoints)
		if err != nil {
			logger.Fatalf("rpc.Init: init etcd resolver failed: %v", err)
		}

		opts := []client.Option{
			client.WithResolver(resolver),
			client.WithRPCTimeout(rpcTimeout),
			client.WithConnectTimeout(connectTimeout),
		}

		// MustNewClient 在初始化失败时 panic：
		// 连不上 etcd 属于启动期配置错误，不该拖到第一个请求才暴露
		userClient = userservice.MustNewClient(cfg.Services.User.Name, opts...)
		videoClient = videoservice.MustNewClient(cfg.Services.Video.Name, opts...)
		commentClient = commentservice.MustNewClient(cfg.Services.Comment.Name, opts...)

		logger.Infof("rpc: clients ready (user=%s video=%s comment=%s)",
			cfg.Services.User.Name, cfg.Services.Video.Name, cfg.Services.Comment.Name)
	})
}

// User 返回用户服务客户端。
func User() userservice.Client { return userClient }

// Video 返回视频服务客户端。
func Video() videoservice.Client { return videoClient }

// Comment 返回评论服务客户端。
func Comment() commentservice.Client { return commentClient }
