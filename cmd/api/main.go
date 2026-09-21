package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"

	"gobili/api/router"
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

	clientset, err := base.NewClientSet(cfg.Snowflake.DatacenterID, cfg.Snowflake.WorkerID)
	if err != nil {
		logger.Fatalf("init clientset failed: %v", err)
	}

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
