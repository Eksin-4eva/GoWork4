package rpc

import (
	"context"

	"gobili/internal/rpcbase"
	"gobili/kitex_gen/video"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// GetFeed 调用视频服务获取最新视频流。
func GetFeed(ctx context.Context, req *video.GetFeedRequest) (*video.GetFeedResponse, error) {
	resp, err := videoClient.GetFeed(ctx, req)
	if err != nil {
		logger.Errorf("rpc.GetFeed: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// PublishVideo 调用视频服务投稿。
func PublishVideo(ctx context.Context, req *video.PublishVideoRequest) error {
	resp, err := videoClient.PublishVideo(ctx, req)
	if err != nil {
		logger.Errorf("rpc.PublishVideo: %v", err)
		return errno.InternalError
	}
	return rpcbase.Check(resp.Base)
}

// GetVideoList 调用视频服务获取用户的投稿列表。
func GetVideoList(ctx context.Context, req *video.GetVideoListRequest) (*video.GetVideoListResponse, error) {
	resp, err := videoClient.GetVideoList(ctx, req)
	if err != nil {
		logger.Errorf("rpc.GetVideoList: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetPopularVideos 调用视频服务获取热门榜。
func GetPopularVideos(ctx context.Context, req *video.GetPopularVideosRequest) (*video.GetPopularVideosResponse, error) {
	resp, err := videoClient.GetPopularVideos(ctx, req)
	if err != nil {
		logger.Errorf("rpc.GetPopularVideos: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// SearchVideos 调用视频服务搜索。
func SearchVideos(ctx context.Context, req *video.SearchVideosRequest) (*video.SearchVideosResponse, error) {
	resp, err := videoClient.SearchVideos(ctx, req)
	if err != nil {
		logger.Errorf("rpc.SearchVideos: %v", err)
		return nil, errno.InternalError
	}
	if err := rpcbase.Check(resp.Base); err != nil {
		return nil, err
	}
	return resp, nil
}

// VisitVideo 调用视频服务记录播放。
func VisitVideo(ctx context.Context, req *video.VisitVideoRequest) error {
	resp, err := videoClient.VisitVideo(ctx, req)
	if err != nil {
		logger.Errorf("rpc.VisitVideo: %v", err)
		return errno.InternalError
	}
	return rpcbase.Check(resp.Base)
}
