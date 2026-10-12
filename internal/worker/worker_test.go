package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gobili/config"
	"gobili/internal/event"
	"gobili/pkg/cache"
	"gobili/pkg/constants"
	"gobili/pkg/db"
	dbmodel "gobili/pkg/db/model"
	"gobili/pkg/kafka"
	"gobili/pkg/utils"
)

const (
	testDSNEnv  = "GOBILI_TEST_DSN"
	testConfig  = "../../config/config.example.yaml"
	testVideoID = int64(9001)
)

// setupDeps 连测试库与本地 Redis。未配置 DSN 时跳过。
func setupDeps(t *testing.T) (*db.Database, *gorm.DB, *cache.Cache) {
	t.Helper()

	dsn := os.Getenv(testDSNEnv)
	if dsn == "" {
		t.Skipf("未设置 %s，跳过 worker 集成测试", testDSNEnv)
	}
	if err := config.Init(testConfig); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	sf, err := utils.NewSnowflake(0, 0)
	if err != nil {
		t.Fatalf("创建雪花生成器失败: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: config.Get().Redis.Addr})
	t.Cleanup(func() { _ = rdb.Close() })

	return db.NewDatabase(gormDB, sf), gormDB, cache.New(rdb)
}

// seedVideo 造一条视频记录。
func seedVideo(t *testing.T, gormDB *gorm.DB, id int64) {
	t.Helper()

	if err := gormDB.Exec("TRUNCATE TABLE " + constants.VideoTableName).Error; err != nil {
		t.Fatalf("清空视频表失败: %v", err)
	}
	video := &dbmodel.Video{ID: id, UserID: 1, VideoKey: "video/x.mp4", Title: "t"}
	if err := gormDB.Table(constants.VideoTableName).Create(video).Error; err != nil {
		t.Fatalf("造视频失败: %v", err)
	}
}

func visitCount(t *testing.T, gormDB *gorm.DB, id int64) int64 {
	t.Helper()

	var video dbmodel.Video
	if err := gormDB.Table(constants.VideoTableName).Where("id = ?", id).First(&video).Error; err != nil {
		t.Fatalf("查询视频失败: %v", err)
	}
	return video.VisitCount
}

// 同一个视频被反复播放时，N 次计数应当合并成 1 次写库。
func TestVisitAggregatorMergesDeltas(t *testing.T) {
	database, gormDB, c := setupDeps(t)
	seedVideo(t, gormDB, testVideoID)

	agg := newVisitAggregator(database, c)

	for i := 0; i < 5; i++ {
		agg.add(testVideoID)
	}
	if got := agg.size(); got != 5 {
		t.Fatalf("缓冲区事件数 = %d, 期望 5", got)
	}

	if err := agg.flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got := visitCount(t, gormDB, testVideoID); got != 5 {
		t.Fatalf("visit_count = %d, 期望 5", got)
	}
	if got := agg.size(); got != 0 {
		t.Fatalf("flush 后缓冲区应为空，实际 %d", got)
	}

	// 空缓冲区 flush 是空操作，不该报错也不该改动数据
	if err := agg.flush(context.Background()); err != nil {
		t.Fatalf("空 flush 不该报错: %v", err)
	}
	if got := visitCount(t, gormDB, testVideoID); got != 5 {
		t.Fatalf("空 flush 后 visit_count 不该变化，实际 %d", got)
	}
}

// 计数天然不幂等，重复投递同一条事件必须被挡住。
func TestConsumeVisitDeduplicatesRedelivery(t *testing.T) {
	database, gormDB, c := setupDeps(t)
	seedVideo(t, gormDB, testVideoID)

	w := newWorker(database, c, nil, nil, stubExtractor{})

	// 事件 ID 每次运行都不同：去重靠 Redis 的 SETNX + TTL，
	// 用固定 ID 的话上一次运行留下的键会让第一次调用就被判成重复，
	// 第二次跑本用例必然失败（这个坑真的踩过一次）。
	eventID := fmt.Sprintf("evt-dedup-%d", time.Now().UnixNano())
	payload, err := json.Marshal(event.VideoVisited{
		V: constants.SchemaVersion, EventID: eventID, VideoID: testVideoID,
	})
	if err != nil {
		t.Fatalf("序列化事件失败: %v", err)
	}

	msg := &kafka.Message{Value: payload}
	w.consumeVisit(context.Background(), msg)
	w.consumeVisit(context.Background(), msg) // Kafka 的重复投递

	if got := w.visits.size(); got != 1 {
		t.Fatalf("重复事件应被去重，缓冲区事件数 = %d, 期望 1", got)
	}
}

// 事件体解析失败只应跳过，不能让消费协程崩掉。
func TestConsumeVisitSkipsMalformedPayload(t *testing.T) {
	database, gormDB, c := setupDeps(t)
	seedVideo(t, gormDB, testVideoID)

	w := newWorker(database, c, nil, nil, stubExtractor{})
	w.consumeVisit(context.Background(), &kafka.Message{Value: []byte("not-json")})

	if got := w.visits.size(); got != 0 {
		t.Fatalf("非法事件不该进缓冲区，实际 %d", got)
	}
}
