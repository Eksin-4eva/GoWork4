// Package cache 封装 Redis 上的业务缓存。
//
// 与 pkg/db 的错误策略不同：这里的失败是**可降级**的。
// Redis 挂掉时业务应当回退到数据库而不是直接失败，
// 所以本包只返回带上下文的原始错误，由调用方决定降级还是上报。
package cache

import "github.com/redis/go-redis/v9"

// Cache 是所有业务缓存的入口。
type Cache struct {
	Popular *Popular
	Dedup   *Dedup
}

// New 构造缓存集合。
func New(rdb *redis.Client) *Cache {
	return &Cache{Popular: NewPopular(rdb), Dedup: NewDedup(rdb)}
}
