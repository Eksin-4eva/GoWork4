package comment

import (
	"context"
	"testing"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

func TestDBComment_CreateTopLevel(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.CommentTableName)

	d := NewDBComment(db, testutil.Snowflake(t))

	// ParentID 用零值 0，绝不能用空字符串——旧版就是这么写的，
	// 空串写进 bigint 列在 MySQL 严格模式下会直接报 1366
	c := &model.Comment{ID: 1, UserID: 1, VideoID: 1, ParentID: 0, Content: "first"}
	if err := d.Create(context.Background(), c); err != nil {
		t.Fatalf("发顶层评论失败: %v", err)
	}

	_, got, err := d.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ParentID != 0 || got.ChildCount != 0 {
		t.Fatalf("顶层评论的 parent_id/child_count 应均为 0，实际 %d/%d", got.ParentID, got.ChildCount)
	}
}

func TestDBComment_CreateReplyBumpsParentChildCount(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.CommentTableName)

	d := NewDBComment(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := d.Create(ctx, &model.Comment{ID: 1, UserID: 1, VideoID: 1, Content: "parent"}); err != nil {
		t.Fatalf("发父评论失败: %v", err)
	}
	for i := int64(0); i < 2; i++ {
		reply := &model.Comment{ID: 2 + i, UserID: 2, VideoID: 1, ParentID: 1, Content: "reply"}
		if err := d.Create(ctx, reply); err != nil {
			t.Fatalf("发回复失败: %v", err)
		}
	}

	_, parent, err := d.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if parent.ChildCount != 2 {
		t.Fatalf("父评论 child_count = %d, 期望 2", parent.ChildCount)
	}
}
