package video

import (
	"context"
	"testing"

	"gobili/pkg/constants"
	"gobili/pkg/db/internal/testutil"
	"gobili/pkg/db/model"
)

func TestDBVideo_Search(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.UserTableName, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))
	ctx := context.Background()

	users := []*model.User{
		{ID: 1, Username: "alice", PasswordHash: "h"},
		{ID: 2, Username: "bob", PasswordHash: "h"},
	}
	if err := db.Table(constants.UserTableName).Create(&users).Error; err != nil {
		t.Fatalf("造用户失败: %v", err)
	}

	videos := []*model.Video{
		{ID: 11, UserID: 1, VideoKey: "v/11.mp4", Title: "golang tutorial", Description: "intro"},
		{ID: 12, UserID: 1, VideoKey: "v/12.mp4", Title: "cooking", Description: "pasta"},
		{ID: 13, UserID: 2, VideoKey: "v/13.mp4", Title: "rust tutorial", Description: "ownership"},
	}
	if err := db.Table(constants.VideoTableName).Create(&videos).Error; err != nil {
		t.Fatalf("造视频失败: %v", err)
	}

	tests := []struct {
		name      string
		params    SearchParams
		wantTotal int64
		wantIDs   []int64
	}{
		{
			name:      "无条件",
			params:    SearchParams{},
			wantTotal: 3,
			wantIDs:   []int64{13, 12, 11},
		},
		{
			name:      "关键词命中标题",
			params:    SearchParams{Keywords: "tutorial"},
			wantTotal: 2,
		},
		{
			name:      "关键词命中简介",
			params:    SearchParams{Keywords: "ownership"},
			wantTotal: 1,
		},
		{
			name:      "按用户名过滤",
			params:    SearchParams{Username: "alice"},
			wantTotal: 2,
		},
		{
			name:      "关键词与用户名组合",
			params:    SearchParams{Keywords: "tutorial", Username: "alice"},
			wantTotal: 1,
			wantIDs:   []int64{11},
		},
		{
			name:      "无命中",
			params:    SearchParams{Keywords: "nonexistent"},
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := d.Search(ctx, tt.params)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if total != tt.wantTotal {
				t.Fatalf("total = %d, 期望 %d", total, tt.wantTotal)
			}
			if len(items) != int(tt.wantTotal) {
				t.Fatalf("返回 %d 条, 期望 %d 条", len(items), tt.wantTotal)
			}
			for i, wantID := range tt.wantIDs {
				if items[i].ID != wantID {
					t.Errorf("第 %d 条 id = %d, 期望 %d", i, items[i].ID, wantID)
				}
			}
		})
	}
}

func TestDBVideo_SearchPagination(t *testing.T) {
	db := testutil.Open(t)
	testutil.Clean(t, db, constants.VideoTableName)

	d := NewDBVideo(db, testutil.Snowflake(t))
	ctx := context.Background()

	for id := int64(1); id <= 5; id++ {
		if err := d.Create(ctx, &model.Video{ID: id, UserID: 1, VideoKey: "v/x.mp4", Title: "t"}); err != nil {
			t.Fatalf("造视频失败: %v", err)
		}
	}

	// total 应是全量 5，而不是当前页的条数
	items, total, err := d.Search(ctx, SearchParams{PageNum: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if total != 5 {
		t.Fatalf("total = %d, 期望 5", total)
	}
	if len(items) != 2 {
		t.Fatalf("第 2 页应返回 2 条, 实际 %d 条", len(items))
	}
}
