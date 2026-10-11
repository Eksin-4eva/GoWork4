package service

import (
	"context"
	"time"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
)

// Feed 返回最新的视频流。from 为 nil 表示不限制起始时间。
//
// 不做分页：feed 是「刷」的语义，客户端每次带上上次最后一条的时间往前推。
func (s *VideoService) Feed(ctx context.Context, from *time.Time) ([]*model.Video, error) {
	return s.db.Video.ListFeed(ctx, from, constants.FeedLimit)
}
