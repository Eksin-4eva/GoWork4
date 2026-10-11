package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// dedupKeyPrefix 是去重键的前缀。
const dedupKeyPrefix = "gobili:dedup:"

// Dedup 用 Redis 对事件去重。
//
// Kafka 是 at-least-once，重复投递是常态。计数类操作天然不幂等
// （+1 执行两次就是多算一次），所以必须能识别出「这条已经处理过了」。
type Dedup struct {
	rdb *redis.Client
}

// NewDedup 构造去重器。
func NewDedup(rdb *redis.Client) *Dedup {
	return &Dedup{rdb: rdb}
}

// FirstSeen 判断某条事件是否为首次出现。
//
// 用 SETNX 把「判断 + 记录」合成一步原子操作：
// 先查再写的话，并发消费时会出现两个协程都认为自己是第一次的窗口。
func (d *Dedup) FirstSeen(ctx context.Context, scope, id string, ttl time.Duration) (bool, error) {
	ok, err := d.rdb.SetNX(ctx, dedupKeyPrefix+scope+":"+id, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("cache.Dedup.FirstSeen: %w", err)
	}
	return ok, nil
}
