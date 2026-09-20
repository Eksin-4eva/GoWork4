package utils

import (
	"fmt"
	"sync"
	"time"
)

// 雪花算法的位分配：
//
//	1 位符号位（固定为 0）+ 41 位时间戳 + 5 位数据中心 + 5 位机器 + 12 位序列号
const (
	epoch            int64 = 1704067200000 // 起始时间：2024-01-01 00:00:00 UTC，可用约 69 年
	timestampBits    int64 = 41
	datacenterIDBits int64 = 5
	workerIDBits     int64 = 5
	sequenceBits     int64 = 12

	timestampMax    int64 = 2199023255551
	datacenterIDMax int64 = 31
	workerIDMax     int64 = 31
	sequenceMask    int64 = 4095

	workerIDShift     int64 = 12
	datacenterIDShift int64 = 17
	timestampShift    int64 = 22

	nanosecondsPerMillisecond = 1_000_000
)

// Snowflake 是并发安全的雪花 ID 生成器。
type Snowflake struct {
	mu           sync.Mutex
	timestamp    int64
	datacenterID int64
	workerID     int64
	sequence     int64
}

// NewSnowflake 创建 ID 生成器，datacenterID 与 workerID 均需落在 [0, 31]。
func NewSnowflake(datacenterID, workerID int64) (*Snowflake, error) {
	if datacenterID < 0 || datacenterID > datacenterIDMax {
		return nil, fmt.Errorf("datacenterID must be between 0 and %d", datacenterIDMax)
	}
	if workerID < 0 || workerID > workerIDMax {
		return nil, fmt.Errorf("workerID must be between 0 and %d", workerIDMax)
	}
	return &Snowflake{datacenterID: datacenterID, workerID: workerID}, nil
}

// NextVal 生成下一个 ID。
//
// 同一毫秒内自增序列号；序列号用尽则自旋等待下一毫秒，
// 保证同一进程内产生的 ID 严格单调且不重复。
func (s *Snowflake) NextVal() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano() / nanosecondsPerMillisecond
	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixNano() / nanosecondsPerMillisecond
			}
		}
	} else {
		s.sequence = 0
	}

	elapsed := now - epoch
	if elapsed > timestampMax {
		return 0, fmt.Errorf("snowflake epoch overflow: %d", elapsed)
	}
	s.timestamp = now

	return elapsed<<timestampShift |
		s.datacenterID<<datacenterIDShift |
		s.workerID<<workerIDShift |
		s.sequence, nil
}

// Timestamp 从 ID 中还原生成时刻。
func Timestamp(id int64) time.Time {
	ms := ((id >> timestampShift) & timestampMax) + epoch
	return time.UnixMilli(ms)
}
