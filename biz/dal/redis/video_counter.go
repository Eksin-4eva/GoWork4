package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/BiliGO/biz/dal/mysql"
	"github.com/BiliGO/biz/dal/query"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	videoVisitCountKey = "video:visit_count:%d"
)

func GetVideoVisitCount(ctx context.Context, videoID int64) (int64, error) {
	key := fmt.Sprintf(videoVisitCountKey, videoID)

	count, err := Client.Get(ctx, key).Int64()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}

func IncrementVideoVisitCount(ctx context.Context, videoID int64) (int64, error) {
	if Client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	key := fmt.Sprintf(videoVisitCountKey, videoID)

	count, err := Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		q := query.Use(mysql.DB)
		v, err := q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).First()
		if err == nil {
			Client.Set(ctx, key, v.VisitCount, 0)
			count = v.VisitCount + 1
		}
	}

	return count, nil
}

func SyncVideoVisitCountToDB(ctx context.Context, videoID int64) error {
	key := fmt.Sprintf(videoVisitCountKey, videoID)

	count, err := Client.Get(ctx, key).Int64()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil
		}
		return err
	}

	q := query.Use(mysql.DB)
	_, err = q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).UpdateSimple(q.Video.VisitCount.Value(count))
	if err != nil {
		return err
	}

	return nil
}

func GetVideoVisitCountWithFallback(ctx context.Context, videoID int64) (int64, error) {
	count, err := GetVideoVisitCount(ctx, videoID)
	if err != nil {
		q := query.Use(mysql.DB)
		v, err := q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).First()
		if err != nil {
			if gorm.ErrRecordNotFound == err {
				return 0, nil
			}
			return 0, err
		}
		return v.VisitCount, nil
	}

	return count, nil
}

func BatchGetVideoVisitCounts(ctx context.Context, videoIDs []int64) (map[int64]int64, error) {
	if Client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	result := make(map[int64]int64)

	pipe := Client.Pipeline()
	cmds := make(map[int64]*redis.StringCmd)

	for _, videoID := range videoIDs {
		key := fmt.Sprintf(videoVisitCountKey, videoID)
		cmds[videoID] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	q := query.Use(mysql.DB)
	missingIDs := make([]int64, 0)

	for videoID, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			if err.Error() == "redis: nil" {
				missingIDs = append(missingIDs, videoID)
				continue
			}
			return nil, err
		}

		count, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}

		result[videoID] = count
	}

	if len(missingIDs) > 0 {
		videos, err := q.Video.WithContext(ctx).Where(q.Video.ID.In(missingIDs...)).Find()
		if err != nil {
			return nil, err
		}

		for _, v := range videos {
			result[v.ID] = v.VisitCount
			key := fmt.Sprintf(videoVisitCountKey, v.ID)
			Client.Set(ctx, key, v.VisitCount, 0)
		}
	}

	return result, nil
}
