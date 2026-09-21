package user

import (
	"context"
	"strings"
	"testing"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
)

func TestDBUser_CreateRejectsDuplicateUsername(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.UserTableName)

	d := NewDBUser(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := d.Create(ctx, &model.User{ID: 1, Username: "alice", PasswordHash: "h"}); err != nil {
		t.Fatalf("首次创建应成功: %v", err)
	}

	err := d.Create(ctx, &model.User{ID: 2, Username: "alice", PasswordHash: "h"})
	if err == nil {
		t.Fatal("重复用户名必须报错")
	}

	// 唯一键冲突应当被翻译成可读的业务错误，
	// 而不是把 "Error 1062: Duplicate entry ..." 直接抛给上层
	got := errno.ConvertErr(err)
	if got.ErrorCode != errno.BizErrorCode {
		t.Fatalf("应落到业务错误码，实际 code=%d msg=%q", got.ErrorCode, got.ErrorMsg)
	}
	if !strings.Contains(got.ErrorMsg, "用户名") {
		t.Fatalf("错误文案应说明是用户名冲突，实际 %q", got.ErrorMsg)
	}
}

// 用户名唯一键不区分软删除，所以删号后用户名仍被占用。
// 这是有意的（防止删号后被人抢注），用测试把它固定下来。
func TestDBUser_CreateKeepsUsernameReservedAfterSoftDelete(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.UserTableName)

	d := NewDBUser(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := d.Create(ctx, &model.User{ID: 1, Username: "bob", PasswordHash: "h"}); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if err := db.Exec("UPDATE users SET deleted_at = NOW() WHERE id = 1").Error; err != nil {
		t.Fatalf("软删失败: %v", err)
	}

	err := d.Create(ctx, &model.User{ID: 2, Username: "bob", PasswordHash: "h"})
	if err == nil {
		t.Fatal("软删后同名用户仍应被拒绝")
	}
}
