package user

import (
	"context"
	"testing"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

// 这条用例同时在验证一件容易踩的事：用 .Table("users") 而不是 .Model() 时，
// GORM 的软删条件到底还会不会自动带上。如果不会，软删的用户会被查出来。
func TestDBUser_GetByIDHidesSoftDeleted(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.UserTableName)

	d := NewDBUser(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := d.Create(ctx, &model.User{ID: 1, Username: "carol", PasswordHash: "h"}); err != nil {
		t.Fatalf("创建失败: %v", err)
	}

	found, u, err := d.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !found || u.Username != "carol" {
		t.Fatalf("应查到 carol，实际 found=%v user=%+v", found, u)
	}

	if err := db.Exec("UPDATE users SET deleted_at = NOW() WHERE id = 1").Error; err != nil {
		t.Fatalf("软删失败: %v", err)
	}

	found, _, err = d.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("软删后 GetByID 不应报错: %v", err)
	}
	if found {
		t.Fatal("软删的用户不该被查出来")
	}
}

func TestDBUser_GetByIDNotFoundIsNotAnError(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.UserTableName)

	d := NewDBUser(db, testutil.Snowflake(t))

	found, u, err := d.GetByID(context.Background(), 999999)
	if err != nil {
		t.Fatalf("查不到记录不该返回错误，实际: %v", err)
	}
	if found || u != nil {
		t.Fatalf("不存在的用户应返回 (false, nil)，实际 found=%v user=%+v", found, u)
	}
}
