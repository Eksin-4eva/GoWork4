package user

import (
	"context"
	"sync"
	"testing"
	"time"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

// seedRefreshToken 往 refresh_tokens 里插一条仍然有效的记录。
func seedRefreshToken(t *testing.T, d *DBUser, id, userID int64, raw string, expiresAt time.Time) {
	t.Helper()

	err := d.SaveRefreshToken(context.Background(), &model.RefreshToken{
		ID:        id,
		UserID:    userID,
		Token:     raw,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatalf("落库刷新令牌失败: %v", err)
	}
}

// revokedOf 读回某条令牌的吊销状态；不存在时返回 false。
func revokedOf(t *testing.T, d *DBUser, raw string) (revoked bool, found bool) {
	t.Helper()

	var row model.RefreshToken
	err := d.client.Table(constants.RefreshTokenTableName).
		Where("token = ?", raw).First(&row).Error
	if err != nil {
		return false, false
	}
	return row.Revoked, true
}

// 换发成功：旧令牌被吊销，新令牌落库。
func TestDBUser_ReplaceRefreshTokenSwaps(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.RefreshTokenTableName)

	d := NewDBUser(db, testutil.Snowflake(t))
	ctx := context.Background()
	seedRefreshToken(t, d, 1, 100, "old-token", time.Now().Add(time.Hour))

	swapped, err := d.ReplaceRefreshToken(ctx, "old-token", 100, &model.RefreshToken{
		ID: 2, UserID: 100, Token: "new-token", ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("ReplaceRefreshToken: %v", err)
	}
	if !swapped {
		t.Fatal("仍然有效的旧令牌应该换发成功")
	}

	if revoked, found := revokedOf(t, d, "old-token"); !found || !revoked {
		t.Fatalf("旧令牌应被吊销且仍在表中（保留审计痕迹），found=%v revoked=%v", found, revoked)
	}
	if revoked, found := revokedOf(t, d, "new-token"); !found || revoked {
		t.Fatalf("新令牌应已落库且未吊销，found=%v revoked=%v", found, revoked)
	}
}

// 同一个旧令牌不能换发两次：第二次必须拿到 false。
func TestDBUser_ReplaceRefreshTokenIsSingleUse(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.RefreshTokenTableName)

	d := NewDBUser(db, testutil.Snowflake(t))
	ctx := context.Background()
	seedRefreshToken(t, d, 1, 100, "old-token", time.Now().Add(time.Hour))

	first, err := d.ReplaceRefreshToken(ctx, "old-token", 100, &model.RefreshToken{
		ID: 2, UserID: 100, Token: "new-a", ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil || !first {
		t.Fatalf("第一次换发应成功: swapped=%v err=%v", first, err)
	}

	second, err := d.ReplaceRefreshToken(ctx, "old-token", 100, &model.RefreshToken{
		ID: 3, UserID: 100, Token: "new-b", ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("第二次换发不该报错（只是换不到）: %v", err)
	}
	if second {
		t.Fatal("已被吊销的旧令牌不该再次换发成功")
	}
}

// 过期、软删、以及不属于该用户的令牌都不能换发。
func TestDBUser_ReplaceRefreshTokenRejectsUnusableTokens(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.RefreshTokenTableName)

	d := NewDBUser(db, testutil.Snowflake(t))
	ctx := context.Background()

	seedRefreshToken(t, d, 1, 100, "expired", time.Now().Add(-time.Minute))
	seedRefreshToken(t, d, 2, 100, "soft-deleted", time.Now().Add(time.Hour))
	seedRefreshToken(t, d, 3, 100, "other-user", time.Now().Add(time.Hour))
	if err := db.Exec("UPDATE refresh_tokens SET deleted_at = NOW() WHERE token = ?",
		"soft-deleted").Error; err != nil {
		t.Fatalf("软删失败: %v", err)
	}

	next := func(id int64) *model.RefreshToken {
		return &model.RefreshToken{ID: id, UserID: 100, Token: "unused", ExpiresAt: time.Now().Add(time.Hour)}
	}

	cases := []struct {
		name string
		old  string
		user int64
	}{
		{"已过期", "expired", 100},
		{"已软删", "soft-deleted", 100},
		{"不属于该用户", "other-user", 999},
		{"根本不存在", "no-such-token", 100},
	}
	for i, c := range cases {
		swapped, err := d.ReplaceRefreshToken(ctx, c.old, c.user, next(int64(10+i)))
		if err != nil {
			t.Fatalf("%s: 不该报错: %v", c.name, err)
		}
		if swapped {
			t.Fatalf("%s 的令牌不该换发成功", c.name)
		}
	}

	// 确认这几次失败没有留下任何新令牌。
	var count int64
	if err := db.Table(constants.RefreshTokenTableName).
		Where("token = ?", "unused").Count(&count).Error; err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if count != 0 {
		t.Fatalf("失败的换发不该写入新令牌，实际写了 %d 条", count)
	}
}

// 并发换发同一个令牌：只能有一个成功。
//
// 这条用例守的是「用原子 SQL 而不是分布式锁」这个设计：
// 有效性写在 UPDATE 的 WHERE 里，数据库保证只有一个并发调用能改到那一行。
func TestDBUser_ReplaceRefreshTokenConcurrent(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.RefreshTokenTableName)

	d := NewDBUser(db, testutil.Snowflake(t))
	ctx := context.Background()
	seedRefreshToken(t, d, 1, 100, "old-token", time.Now().Add(time.Hour))

	const workers = 8
	results := make([]bool, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			swapped, err := d.ReplaceRefreshToken(ctx, "old-token", 100, &model.RefreshToken{
				ID:        int64(100 + i),
				UserID:    100,
				Token:     "concurrent-" + string(rune('a'+i)),
				ExpiresAt: time.Now().Add(time.Hour),
			})
			if err != nil {
				t.Errorf("并发换发出错: %v", err)
				return
			}
			results[i] = swapped
		}(i)
	}
	wg.Wait()

	success := 0
	for _, ok := range results {
		if ok {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("并发换发应恰好成功 1 次，实际 %d 次", success)
	}
}
