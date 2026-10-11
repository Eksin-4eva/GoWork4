package service

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
)

// List 分页查询评论。
//
// parentID 为 0 表示查视频的顶层评论，非 0 表示查该评论下的回复。
func (s *CommentService) List(
	ctx context.Context, videoID, parentID int64, pageNum, pageSize int,
) ([]*model.Comment, int64, error) {
	if videoID <= 0 {
		return nil, 0, errno.ParamError.WithMessage("video_id 无效")
	}
	return s.db.Comment.ListByVideo(ctx, videoID, parentID, pageNum, pageSize)
}
