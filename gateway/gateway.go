// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"gobili/gateway/internal/config"
	"gobili/gateway/internal/handler"
	"gobili/gateway/internal/svc"
	"gobili/pkg/migration"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := migration.RunMigrationWithRetry(c.MySQL.DSN, 5, 3*time.Second); err != nil {
		fmt.Printf("Database migration failed: %v\n", err)
	} else {
		fmt.Println("Database migration completed successfully")
	}

	server := rest.MustNewServer(c.RestConf, rest.WithFileServer("/static", http.Dir("uploads")))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
