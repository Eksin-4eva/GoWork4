package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"gobili/internal/event"
	"gobili/pkg/cache"
	"gobili/pkg/constants"
	"gobili/pkg/db"
	"gobili/pkg/kafka"
	"gobili/pkg/logger"
)

// flushTimeout 是退出前最后一次提交的超时。
const flushTimeout = 10 * time.Second

// visitAggregator 在内存里聚合播放计数，定期批量提交。
//
// 同一视频被反复播放时，N 次写库会合并成 1 次——这正是把计数异步化的收益所在。
type visitAggregator struct {
	db    *db.Database
	cache *cache.Cache

	mu      sync.Mutex
	pending map[int64]int64
	count   int64 // 缓冲区累计的事件数，避免每次判空都遍历 map
}

func newVisitAggregator(database *db.Database, c *cache.Cache) *visitAggregator {
	return &visitAggregator{
		db:      database,
		cache:   c,
		pending: make(map[int64]int64),
	}
}

// add 累加一个视频的增量。
func (a *visitAggregator) add(videoID int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pending[videoID]++
	a.count++
}

// size 返回缓冲区里累计的事件数。
func (a *visitAggregator) size() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.count
}

// flush 把缓冲区提交到数据库与热榜。
//
// 先把缓冲换出来再提交：提交期间新到的事件进新的 map，
// 不会因为这一次提交失败而跟着一起丢。
func (a *visitAggregator) flush(ctx context.Context) error {
	a.mu.Lock()
	if len(a.pending) == 0 {
		a.mu.Unlock()
		return nil
	}
	deltas := a.pending
	a.pending = make(map[int64]int64)
	a.count = 0
	a.mu.Unlock()

	if err := a.db.Video.BatchIncrementVisit(ctx, deltas); err != nil {
		a.restore(deltas)
		return err
	}

	for videoID, delta := range deltas {
		if err := a.cache.Popular.Incr(ctx, videoID, delta); err != nil {
			logger.Warnf("worker: 更新热榜失败 video=%d: %v", videoID, err)
		}
	}
	return nil
}

// restore 把提交失败的增量放回缓冲区，等下次 flush 再试。
func (a *visitAggregator) restore(deltas map[int64]int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for videoID, delta := range deltas {
		a.pending[videoID] += delta
		a.count += delta
	}
}

// RunVisitAggregator 消费播放事件并定期 flush，直到 ctx 结束。
func (w *Worker) RunVisitAggregator(ctx context.Context) {
	messages := w.mq.Consumer.Start(ctx, constants.TopicVideoVisited, constants.VisitedConsumerNum)

	// 退出前把缓冲区里剩下的计数提交掉。
	// 少了这一步，最后几秒的播放数会随进程一起消失。
	defer func() {
		finalCtx, cancel := context.WithTimeout(context.Background(), flushTimeout)
		defer cancel()
		if err := w.visits.flush(finalCtx); err != nil {
			logger.Errorf("worker: 退出前提交计数失败: %v", err)
		}
	}()

	ticker := time.NewTicker(constants.VisitFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				return
			}
			w.consumeVisit(ctx, msg)
		case <-ticker.C:
			if err := w.visits.flush(ctx); err != nil {
				logger.Errorf("worker: 提交播放计数失败: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// consumeVisit 校验并累加一条播放事件。
func (w *Worker) consumeVisit(ctx context.Context, msg *kafka.Message) {
	var ev event.VideoVisited
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		logger.Errorf("worker: 解析播放事件失败，跳过: %v", err)
		return
	}
	if ev.VideoID <= 0 {
		return
	}

	// 计数天然不幂等（+1 执行两次就是多算一次），用事件 ID 挡住重复投递。
	// Redis 不可用时按「首次」处理：宁可多算一点，也不要静默漏算。
	if ev.EventID != "" {
		first, err := w.cache.Dedup.FirstSeen(ctx, "visit", ev.EventID, constants.DedupTTL)
		if err != nil {
			logger.Warnf("worker: 去重查询失败，按首次处理: %v", err)
		} else if !first {
			return
		}
	}

	w.visits.add(ev.VideoID)

	// 攒够一批就立刻提交，不必等定时器
	if w.visits.size() >= constants.VisitFlushSize {
		if err := w.visits.flush(ctx); err != nil {
			logger.Errorf("worker: 提交播放计数失败: %v", err)
		}
	}
}
