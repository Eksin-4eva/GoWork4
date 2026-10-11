// Package comment 实现 Kitex 生成的 CommentService 接口。
package comment

import (
	"context"

	"gobili/internal/comment/pack"
	"gobili/internal/comment/service"
	"gobili/internal/rpcbase"
	"gobili/kitex_gen/comment"
)

// CommentServiceImpl 实现 kitex_gen/comment 里的 CommentService 接口。
type CommentServiceImpl struct {
	svc *service.CommentService
}

// NewCommentServiceImpl 构造评论服务实现。
func NewCommentServiceImpl() *CommentServiceImpl {
	return &CommentServiceImpl{svc: service.New()}
}

// PublishComment 发表评论或回复。
func (s *CommentServiceImpl) PublishComment(
	ctx context.Context, req *comment.PublishCommentRequest,
) (*comment.PublishCommentResponse, error) {
	parentID := int64(0)
	if req.ParentId != nil {
		parentID = *req.ParentId
	}

	err := s.svc.Publish(ctx, req.UserId, req.VideoId, parentID, req.Content)
	if err != nil {
		return &comment.PublishCommentResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &comment.PublishCommentResponse{Base: rpcbase.OK()}, nil
}

// GetCommentList 获取评论列表。
func (s *CommentServiceImpl) GetCommentList(
	ctx context.Context, req *comment.GetCommentListRequest,
) (*comment.GetCommentListResponse, error) {
	parentID := int64(0)
	if req.ParentId != nil {
		parentID = *req.ParentId
	}

	items, total, err := s.svc.List(ctx, req.VideoId, parentID,
		int(derefInt64(req.PageNum)), int(derefInt64(req.PageSize)))
	if err != nil {
		return &comment.GetCommentListResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &comment.GetCommentListResponse{
		Base:  rpcbase.OK(),
		Items: pack.BuildComments(items),
		Total: &total,
	}, nil
}

// DeleteComment 删除自己的评论。
func (s *CommentServiceImpl) DeleteComment(
	ctx context.Context, req *comment.DeleteCommentRequest,
) (*comment.DeleteCommentResponse, error) {
	if err := s.svc.Delete(ctx, req.UserId, req.CommentId); err != nil {
		return &comment.DeleteCommentResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &comment.DeleteCommentResponse{Base: rpcbase.OK()}, nil
}

// derefInt64 解引用可选的整型字段。
func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
