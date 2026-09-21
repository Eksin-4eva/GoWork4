package video

import (
	"context"
	"testing"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

func TestDBVideo_IncrementVisit(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := d.Create(ctx, &model.Video{ID: 1, UserID: 1, VideoKey: "video/1.mp4", Title: "t"}); err != nil {
		t.Fatalf("创建视频失败: %v", err)
	}

	ok, err := d.IncrementVisit(ctx, 1)
	if err != nil {
		t.Fatalf("IncrementVisit: %v", err)
	}
	if !ok {
		t.Fatal("存在的视频应返回 true")
	}

	// 连加三次，计数应为 3
	for i := 0; i < 2; i++ {
		if _, err := d.IncrementVisit(ctx, 1); err != nil {
			t.Fatalf("IncrementVisit: %v", err)
		}
	}

	_, v, err := d.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if v.VisitCount != 3 {
		t.Fatalf("visit_count = %d, 期望 3", v.VisitCount)
	}
}

// 旧版是靠「先查视频再更新」做存在性校验的，异步化之后这个校验必须由
// UPDATE 的 RowsAffected 来承担。这条用例把该行为固定下来。
func TestDBVideo_IncrementVisitReportsMissingVideo(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))

	ok, err := d.IncrementVisit(context.Background(), 123456)
	if err != nil {
		t.Fatalf("不存在的视频不该返回错误: %v", err)
	}
	if ok {
		t.Fatal("不存在的视频应返回 false")
	}
}

func TestDBVideo_IncrementVisitIgnoresSoftDeleted(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := d.Create(ctx, &model.Video{ID: 2, UserID: 1, VideoKey: "video/2.mp4", Title: "t"}); err != nil {
		t.Fatalf("创建视频失败: %v", err)
	}
	if err := db.Exec("UPDATE videos SET deleted_at = NOW() WHERE id = 2").Error; err != nil {
		t.Fatalf("软删失败: %v", err)
	}

	ok, err := d.IncrementVisit(ctx, 2)
	if err != nil {
		t.Fatalf("IncrementVisit: %v", err)
	}
	if ok {
		t.Fatal("软删的视频不该被计数")
	}
}
