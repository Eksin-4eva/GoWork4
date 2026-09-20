package client

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"gobili/config"
)

// redisPingTimeout 是初始化时探测 Redis 连通性的超时。
const redisPingTimeout = 5 * time.Second

// InitRedis 初始化 Redis 客户端并做连通性检查。
func InitRedis() (*redis.Client, error) {
	c := config.Get().Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("client.InitRedis: ping redis: %w", err)
	}
	return rdb, nil
}
