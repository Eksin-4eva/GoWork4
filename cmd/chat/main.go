// cmd/chat 启动 WebSocket 聊天服务。
//
// 与 api 分开部署：聊天连接是长连接，扩缩容节奏和 HTTP 网关完全不同，
// 拆开之后各自的发布互不影响。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gobili/config"
	"gobili/internal/chat/service"
	"gobili/pkg/base"
	"gobili/pkg/constants"
	"gobili/pkg/logger"
)

// shutdownTimeout 是收到退出信号后等待在途请求结束的时间。
const shutdownTimeout = 10 * time.Second

var configFile = flag.String("f", "config/config.yaml", "the config file path")

func main() {
	flag.Parse()

	if err := config.Init(*configFile); err != nil {
		panic(err)
	}
	cfg := config.Get()
	logger.Init(constants.ServiceChat, cfg.Server.LogLevel)

	// 聊天服务只需要数据库：消息落库、刷新令牌校验
	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID,
		base.WithMySQL())
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}
	defer clientset.Close()

	svc := service.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	mux.HandleFunc("/history", svc.HandleHistory)
	mux.HandleFunc("/conversations", svc.HandleConversations)
	mux.HandleFunc("/healthz", svc.HandleHealth)

	addr := fmt.Sprintf("%s:%d", cfg.Chat.Host, cfg.Chat.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
		// 慢连接不应该把服务端连接池占满
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Infof("%s listening on %s", constants.ServiceChat, addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("chat server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("%s shutting down", constants.ServiceChat)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("chat shutdown: %v", err)
	}
}
