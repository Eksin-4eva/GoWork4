package utils

import (
	"sync"
	"testing"
	"time"
)

func TestNewSnowflakeRejectsOutOfRangeIDs(t *testing.T) {
	if _, err := NewSnowflake(-1, 0); err == nil {
		t.Fatal("NewSnowflake(-1, 0) should fail")
	}
	if _, err := NewSnowflake(0, 32); err == nil {
		t.Fatal("NewSnowflake(0, 32) should fail")
	}
}

func TestNextValIsUnique(t *testing.T) {
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("NewSnowflake: %v", err)
	}

	const n = 20000
	seen := make(map[int64]struct{}, n)
	for i := 0; i < n; i++ {
		id, err := sf.NextVal()
		if err != nil {
			t.Fatalf("NextVal: %v", err)
		}
		if id <= 0 {
			t.Fatalf("NextVal returned non-positive id %d", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("NextVal returned duplicate id %d at iteration %d", id, i)
		}
		seen[id] = struct{}{}
	}
}

func TestNextValIsMonotonic(t *testing.T) {
	sf, err := NewSnowflake(0, 0)
	if err != nil {
		t.Fatalf("NewSnowflake: %v", err)
	}

	prev, err := sf.NextVal()
	if err != nil {
		t.Fatalf("NextVal: %v", err)
	}
	for i := 0; i < 5000; i++ {
		cur, err := sf.NextVal()
		if err != nil {
			t.Fatalf("NextVal: %v", err)
		}
		if cur <= prev {
			t.Fatalf("id not monotonic: %d then %d", prev, cur)
		}
		prev = cur
	}
}

func TestTimestampRoundTrip(t *testing.T) {
	sf, err := NewSnowflake(0, 0)
	if err != nil {
		t.Fatalf("NewSnowflake: %v", err)
	}

	before := time.Now().Add(-time.Second)
	id, err := sf.NextVal()
	if err != nil {
		t.Fatalf("NextVal: %v", err)
	}
	after := time.Now().Add(time.Second)

	got := Timestamp(id)
	if got.Before(before) || got.After(after) {
		t.Fatalf("Timestamp(id) = %v, want within [%v, %v]", got, before, after)
	}
}

// 并发才是雪花 ID 的真实风险点：timestamp 与 sequence 由同一把互斥锁保护，
// 串行测试测不出锁写错。这条配合 `make test` 的 -race 才有意义。
func TestNextValIsUniqueUnderConcurrency(t *testing.T) {
	sf, err := NewSnowflake(3, 7)
	if err != nil {
		t.Fatalf("NewSnowflake: %v", err)
	}

	const (
		goroutines = 16
		perRoutine = 5000
		total      = goroutines * perRoutine
	)

	var (
		mu   sync.Mutex
		seen = make(map[int64]struct{}, total)
		wg   sync.WaitGroup
	)

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perRoutine; i++ {
				id, err := sf.NextVal()
				if err != nil {
					t.Errorf("NextVal: %v", err)
					return
				}
				mu.Lock()
				if _, dup := seen[id]; dup {
					t.Errorf("duplicate id %d", id)
				}
				seen[id] = struct{}{}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(seen) != total {
		t.Fatalf("got %d unique ids, want %d", len(seen), total)
	}
}

// NextVal 里最难覆盖的分支：同一毫秒内 12 位序列号用尽，必须自旋等到下一毫秒。
// 只有「生成 4096 个 ID 快于 1 毫秒」时才会触发，所以循环到观察到为止；
// 机器慢到始终触发不了时跳过，而不是误报失败。
func TestNextValSequenceOverflowAdvancesTimestamp(t *testing.T) {
	sf, err := NewSnowflake(0, 0)
	if err != nil {
		t.Fatalf("NewSnowflake: %v", err)
	}

	prev, err := sf.NextVal()
	if err != nil {
		t.Fatalf("NextVal: %v", err)
	}

	const (
		wantOverflows = 3
		maxIterations = 500_000
	)

	overflows := 0
	for i := 0; i < maxIterations && overflows < wantOverflows; i++ {
		cur, err := sf.NextVal()
		if err != nil {
			t.Fatalf("NextVal: %v", err)
		}
		if cur <= prev {
			t.Fatalf("id not monotonic: %d then %d", prev, cur)
		}
		// 序列号回绕，说明刚才跨过了"同一毫秒内序列用尽"那一步
		if cur&sequenceMask <= prev&sequenceMask {
			overflows++
			// 回绕的同时时间戳必须前进，否则会与本毫秒早先发出的 ID 重复
			if cur>>timestampShift <= prev>>timestampShift {
				t.Fatalf("序列号回绕但时间戳未前进: prev=%d cur=%d", prev, cur)
			}
		}
		prev = cur
	}

	if overflows < wantOverflows {
		t.Skipf("本机生成 4096 个 ID 慢于 1ms，未触发序列号回绕分支（已观察 %d 次）", overflows)
	}
}
