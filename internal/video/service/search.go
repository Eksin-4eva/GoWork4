package service

import (
	"context"
	"strings"
	"time"

	"gobili/pkg/db/model"
	dbvideo "gobili/pkg/db/video"
)

// SearchParams 是搜索入参，零值字段表示不加该条件。
type SearchParams struct {
	Keywords string
	Username string
	FromTime *time.Time
	ToTime   *time.Time
	PageNum  int
	PageSize int
}

// Search 按关键词 / 投稿人 / 时间区间组合搜索视频。
func (s *VideoService) Search(ctx context.Context, p SearchParams) ([]*model.Video, int64, error) {
	return s.db.Video.Search(ctx, dbvideo.SearchParams{
		Keywords: strings.TrimSpace(p.Keywords),
		Username: strings.TrimSpace(p.Username),
		FromTime: p.FromTime,
		ToTime:   p.ToTime,
		PageNum:  p.PageNum,
		PageSize: p.PageSize,
	})
}
