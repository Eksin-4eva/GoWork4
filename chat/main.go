package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"gobili/chat/internal/config"
	"gobili/chat/internal/server"
	"gobili/chat/internal/service"
	"gobili/pkg/migration"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/chat.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := migration.RunMigrationWithRetry(c.MySQL.DSN, 5, 3*time.Second); err != nil {
		fmt.Printf("Database migration failed: %v\n", err)
	} else {
		fmt.Println("Database migration completed successfully")
	}

	svc, err := service.New(c)
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Close()

	srv := server.New(c, svc)
	fmt.Printf("Starting chat server at %s:%d...\n", c.Host, c.Port)
	log.Fatal(srv.ListenAndServe())
}
