package comment

import (
	"context"
	"testing"

	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

// assertCounters 校验视频评论数与父评论回复数。
func assertCounters(t *testing.T, db *gorm.DB, wantVideoComment, wantParentChild int64) {
	t.Helper()

	var v model.Video
	if err := db.Table(constants.VideoTableName).Where("id = ?", 1).First(&v).Error; err != nil {
		t.Fatalf("查视频失败: %v", err)
	}
	if v.CommentCount != wantVideoComment {
		t.Errorf("视频 comment_count = %d, 期望 %d", v.CommentCount, wantVideoComment)
	}

	var parent model.Comment
	if err := db.Table(constants.CommentTableName).Where("id = ?", 1).First(&parent).Error; err != nil {
		t.Fatalf("查父评论失败: %v", err)
	}
	if parent.ChildCount != wantParentChild {
		t.Errorf("父评论 child_count = %d, 期望 %d", parent.ChildCount, wantParentChild)
	}
}

func TestDBComment_SoftDeleteRollsBackCounters(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.CommentTableName, constants.VideoTableName)

	d := NewDBComment(db, testutil.Snowflake(t))
	ctx := context.Background()

	if err := db.Table(constants.VideoTableName).
		Create(&model.Video{ID: 1, UserID: 1, VideoKey: "v/1.mp4", Title: "t"}).Error; err != nil {
		t.Fatalf("造视频失败: %v", err)
	}
	if err := d.Create(ctx, &model.Comment{ID: 1, UserID: 1, VideoID: 1, Content: "parent"}); err != nil {
		t.Fatalf("发父评论失败: %v", err)
	}
	if err := d.Create(ctx, &model.Comment{ID: 2, UserID: 2, VideoID: 1, ParentID: 1, Content: "reply"}); err != nil {
		t.Fatalf("发回复失败: %v", err)
	}

	// 两条评论都计入视频总数，只有回复计入父评论的 child_count
	assertCounters(t, db, 2, 1)

	deleted, err := d.SoftDelete(ctx, 2, 1, 1)
	if err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	if !deleted {
		t.Fatal("首次删除应生效")
	}
	assertCounters(t, db, 1, 0)

	// 重复删除要被幂等护栏拦下，且计数不能再减第二次
	deleted, err = d.SoftDelete(ctx, 2, 1, 1)
	if err != nil {
		t.Fatalf("重复 SoftDelete: %v", err)
	}
	if deleted {
		t.Fatal("重复删除应返回 false")
	}
	assertCounters(t, db, 1, 0)
}

func TestDBComment_SoftDeleteMissingComment(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.CommentTableName, constants.VideoTableName)

	d := NewDBComment(db, testutil.Snowflake(t))

	deleted, err := d.SoftDelete(context.Background(), 999, 1, 0)
	if err != nil {
		t.Fatalf("删除不存在的评论不该报错: %v", err)
	}
	if deleted {
		t.Fatal("删除不存在的评论应返回 false")
	}
}
