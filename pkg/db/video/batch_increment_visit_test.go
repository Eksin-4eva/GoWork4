package video

import (
	"context"
	"testing"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

func TestDBVideo_BatchIncrementVisit(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))
	ctx := context.Background()

	for _, id := range []int64{1, 2, 3} {
		if err := d.Create(ctx, &model.Video{ID: id, UserID: 1, VideoKey: "video/x.mp4", Title: "t"}); err != nil {
			t.Fatalf("创建视频失败: %v", err)
		}
	}

	// worker 一次 flush 把多个视频的累计增量一起提交
	if err := d.BatchIncrementVisit(ctx, map[int64]int64{1: 5, 2: 1, 3: 0}); err != nil {
		t.Fatalf("BatchIncrementVisit: %v", err)
	}

	want := map[int64]int64{1: 5, 2: 1, 3: 0}
	for id, expect := range want {
		_, v, err := d.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("GetByID(%d): %v", id, err)
		}
		if v.VisitCount != expect {
			t.Errorf("视频 %d 的 visit_count = %d, 期望 %d", id, v.VisitCount, expect)
		}
	}
}

func TestDBVideo_BatchIncrementVisitEmptyIsNoop(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))

	if err := d.BatchIncrementVisit(context.Background(), nil); err != nil {
		t.Fatalf("空批次应是空操作，实际报错: %v", err)
	}
}
