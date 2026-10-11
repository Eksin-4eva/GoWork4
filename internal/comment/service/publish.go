package service

import (
	"context"
	"strings"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Publish 发表评论（parentID 为 0）或回复某条评论（parentID 非 0）。
func (s *CommentService) Publish(
	ctx context.Context, userID, videoID, parentID int64, content string,
) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errno.ParamEmpty.WithMessage("评论内容不能为空")
	}
	if len([]rune(content)) > constants.MaxCommentRunes {
		return errno.ParamError.WithMessage("评论不能超过 500 字")
	}

	// 视频必须真实存在，否则评论区会挂在一条不存在的视频上
	found, _, err := s.db.Video.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if !found {
		return errno.BizNotExist.WithMessage("视频不存在")
	}

	if parentID != 0 {
		found, parent, err := s.db.Comment.GetByID(ctx, parentID)
		if err != nil {
			return err
		}
		if !found {
			return errno.BizNotExist.WithMessage("要回复的评论不存在")
		}
		// 不允许把 A 视频的评论挂在 B 视频下面，
		// 否则评论列表与回复计数会对不上
		if parent.VideoID != videoID {
			return errno.ParamError.WithMessage("评论与视频不匹配")
		}
	}

	id, err := s.sf.NextVal()
	if err != nil {
		logger.Errorf("comment.Publish: snowflake: %v", err)
		return errno.InternalError
	}

	return s.db.Comment.Create(ctx, &model.Comment{
		ID:       id,
		UserID:   userID,
		VideoID:  videoID,
		ParentID: parentID,
		Content:  content,
	})
}
