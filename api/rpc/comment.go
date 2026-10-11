package rpc

import (
	"context"

	"gobili/internal/rpcbase"
	"gobili/kitex_gen/comment"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// PublishComment 调用评论服务发表评论。
func PublishComment(ctx context.Context, req *comment.PublishCommentRequest) error {
	resp, err := commentClient.PublishComment(ctx, req)
	if err != nil {
		logger.Errorf("rpc.PublishComment: %v", err)
		return errno.InternalError
	}
	return rpcbase.Check(resp.Base)
}

// GetCommentList 调用评论服务获取评论列表。
func GetCommentList(ctx context.Context, req *comment.GetCommentListRequest) (*comment.GetCommentListResponse, error) {
	resp, err := commentClient.GetCommentList(ctx, req)
	if err != nil {
		logger.Errorf("rpc.GetCommentList: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteComment 调用评论服务删除评论。
func DeleteComment(ctx context.Context, req *comment.DeleteCommentRequest) error {
	resp, err := commentClient.DeleteComment(ctx, req)
	if err != nil {
		logger.Errorf("rpc.DeleteComment: %v", err)
		return errno.InternalError
	}
	return rpcbase.Check(resp.Base)
}
