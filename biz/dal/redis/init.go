package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

func GetConfigFromEnv() Config {
	return Config{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
}

func Init(cfg Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	log.Println("redis initialized successfully")
	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
