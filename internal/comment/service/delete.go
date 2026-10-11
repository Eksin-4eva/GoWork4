package service

import (
	"context"

	"gobili/pkg/errno"
)

// Delete 删除自己的评论。
//
// 权限判断放在业务层：作者不匹配时返回的是「无权限」，
// 与「评论不存在」是不同的语义，但对外都不透露评论的实际归属。
func (s *CommentService) Delete(ctx context.Context, userID, commentID int64) error {
	found, c, err := s.db.Comment.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if !found {
		return errno.BizNotExist.WithMessage("评论不存在")
	}
	if c.UserID != userID {
		return errno.BizForbidden.WithMessage("只能删除自己的评论")
	}

	// 并发重复删除时只有第一次会真正生效，这里把它当成幂等成功
	if _, err := s.db.Comment.SoftDelete(ctx, c.ID, c.VideoID, c.ParentID); err != nil {
		return err
	}
	return nil
}
