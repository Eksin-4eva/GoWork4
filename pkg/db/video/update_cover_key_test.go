package video

import (
	"context"
	"testing"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

// worker 会因 Kafka 至少一次投递而重复处理同一条消息，
// 所以「回写封面」必须天然幂等：第二次写入要被丢弃，且不能覆盖已有值。
func TestDBVideo_SetCoverKeyIsIdempotent(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := d.Create(ctx, &model.Video{ID: 1, UserID: 1, VideoKey: "video/1.mp4", Title: "t"}); err != nil {
		t.Fatalf("创建视频失败: %v", err)
	}

	ok, err := d.SetCoverKey(ctx, 1, "cover/1.jpg")
	if err != nil {
		t.Fatalf("首次回写: %v", err)
	}
	if !ok {
		t.Fatal("首次回写应生效")
	}

	ok, err = d.SetCoverKey(ctx, 1, "cover/1-should-not-win.jpg")
	if err != nil {
		t.Fatalf("重复回写: %v", err)
	}
	if ok {
		t.Fatal("重复回写应被幂等护栏拦下")
	}

	_, v, err := d.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if v.CoverKey != "cover/1.jpg" {
		t.Fatalf("cover_key = %q, 期望保持首次写入的值", v.CoverKey)
	}
}

func TestDBVideo_ListMissingCover(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))
	ctx := context.Background()

	for _, id := range []int64{1, 2} {
		if err := d.Create(ctx, &model.Video{ID: id, UserID: 1, VideoKey: "video/x.mp4", Title: "t"}); err != nil {
			t.Fatalf("创建视频失败: %v", err)
		}
	}
	if _, err := d.SetCoverKey(ctx, 1, "cover/1.jpg"); err != nil {
		t.Fatalf("回写封面失败: %v", err)
	}

	items, err := d.ListMissingCover(ctx, 10)
	if err != nil {
		t.Fatalf("ListMissingCover: %v", err)
	}
	if len(items) != 1 || items[0].ID != 2 {
		t.Fatalf("应只剩 id=2 缺封面，实际 %+v", items)
	}
}
