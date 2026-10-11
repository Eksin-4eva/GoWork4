package service

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
)

// List 分页查询某个用户的投稿。
func (s *VideoService) List(
	ctx context.Context, userID int64, pageNum, pageSize int,
) ([]*model.Video, int64, error) {
	if userID <= 0 {
		return nil, 0, errno.ParamError.WithMessage("user_id 无效")
	}
	return s.db.Video.ListByUser(ctx, userID, pageNum, pageSize)
}
