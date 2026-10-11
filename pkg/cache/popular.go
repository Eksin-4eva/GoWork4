package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// 热门榜的 Redis 键。
const (
	// popularKey 是有序集合：member 是视频 ID，score 是播放量。
	popularKey = "gobili:video:popular"
	// popularReadyKey 标记榜单是否已经初始化过。
	//
	// 这个标记是必需的：Redis 冷启动时集合就是空的，
	// 只看「集合有没有数据」会把「尚未初始化」误判成「所有视频都是 0 播放」，
	// 于是热门榜直接返回空。
	popularReadyKey = "gobili:video:popular:ready"
	// popularReadyTTL 到期后会用数据库重新校正一次榜单，
	// 顺带修正长时间累积的分数漂移。
	popularReadyTTL = 10 * time.Minute
)

// Popular 用有序集合维护「按播放量排序的视频 ID」。
type Popular struct {
	rdb *redis.Client
}

// NewPopular 构造热门榜缓存。
func NewPopular(rdb *redis.Client) *Popular {
	return &Popular{rdb: rdb}
}

// Incr 给某个视频的分数加 delta。
func (p *Popular) Incr(ctx context.Context, videoID, delta int64) error {
	if err := p.rdb.ZIncrBy(ctx, popularKey, float64(delta), strconv.FormatInt(videoID, 10)).Err(); err != nil {
		return fmt.Errorf("cache.Popular.Incr: %w", err)
	}
	return nil
}

// TopIDs 按分数从高到低取一段视频 ID。
func (p *Popular) TopIDs(ctx context.Context, offset, count int64) ([]int64, error) {
	// 用 ZRangeArgs(Rev) 而不是已废弃的 ZRevRange
	members, err := p.rdb.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   popularKey,
		Start: offset,
		Stop:  offset + count - 1,
		Rev:   true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("cache.Popular.TopIDs: %w", err)
	}
	return parseIDs(members)
}

// Count 返回榜单里的视频总数。
func (p *Popular) Count(ctx context.Context) (int64, error) {
	total, err := p.rdb.ZCard(ctx, popularKey).Result()
	if err != nil {
		return 0, fmt.Errorf("cache.Popular.Count: %w", err)
	}
	return total, nil
}

// IsReady 判断榜单是否已经初始化过。
func (p *Popular) IsReady(ctx context.Context) (bool, error) {
	exists, err := p.rdb.Exists(ctx, popularReadyKey).Result()
	if err != nil {
		return false, fmt.Errorf("cache.Popular.IsReady: %w", err)
	}
	return exists > 0, nil
}

// Replace 用一批分数重建整个榜单，并标记为已就绪。
//
// 用事务管道：先删旧集合再写新数据，中间不会被其他请求读到半截数据。
func (p *Popular) Replace(ctx context.Context, scores map[int64]int64) error {
	pipe := p.rdb.TxPipeline()
	pipe.Del(ctx, popularKey)

	if len(scores) > 0 {
		members := make([]redis.Z, 0, len(scores))
		for id, score := range scores {
			members = append(members, redis.Z{
				Score:  float64(score),
				Member: strconv.FormatInt(id, 10),
			})
		}
		pipe.ZAdd(ctx, popularKey, members...)
	}
	pipe.Set(ctx, popularReadyKey, "1", popularReadyTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("cache.Popular.Replace: %w", err)
	}
	return nil
}

// parseIDs 把集合成员转成视频 ID。
func parseIDs(members []string) ([]int64, error) {
	ids := make([]int64, 0, len(members))
	for _, member := range members {
		id, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cache: 榜单成员 %q 不是合法的视频 ID: %w", member, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
