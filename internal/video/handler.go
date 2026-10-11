// Package video 实现 Kitex 生成的 VideoService 接口。
package video

import (
	"context"
	"time"

	"gobili/internal/rpcbase"
	"gobili/internal/video/pack"
	"gobili/internal/video/service"
	"gobili/kitex_gen/video"
)

// VideoServiceImpl 实现 kitex_gen/video 里的 VideoService 接口。
type VideoServiceImpl struct {
	svc *service.VideoService
}

// NewVideoServiceImpl 构造视频服务实现。
func NewVideoServiceImpl() *VideoServiceImpl {
	return &VideoServiceImpl{svc: service.New()}
}

// GetFeed 获取最新视频流。
func (s *VideoServiceImpl) GetFeed(
	ctx context.Context, req *video.GetFeedRequest,
) (*video.GetFeedResponse, error) {
	items, err := s.svc.Feed(ctx, millisToTime(req.FromMillis))
	if err != nil {
		return &video.GetFeedResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &video.GetFeedResponse{Base: rpcbase.OK(), Items: pack.BuildVideos(items)}, nil
}

// PublishVideo 投稿视频。
func (s *VideoServiceImpl) PublishVideo(
	ctx context.Context, req *video.PublishVideoRequest,
) (*video.PublishVideoResponse, error) {
	err := s.svc.Publish(ctx, req.UserId, service.PublishVideoParams{
		Title:       req.Title,
		Description: req.Description,
		VideoKey:    req.VideoKey,
		CoverKey:    derefString(req.CoverKey),
	})
	if err != nil {
		return &video.PublishVideoResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &video.PublishVideoResponse{Base: rpcbase.OK()}, nil
}

// GetVideoList 获取指定用户的投稿列表。
func (s *VideoServiceImpl) GetVideoList(
	ctx context.Context, req *video.GetVideoListRequest,
) (*video.GetVideoListResponse, error) {
	items, total, err := s.svc.List(ctx, req.UserId,
		int(derefInt64(req.PageNum)), int(derefInt64(req.PageSize)))
	if err != nil {
		return &video.GetVideoListResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &video.GetVideoListResponse{
		Base:  rpcbase.OK(),
		Items: pack.BuildVideos(items),
		Total: &total,
	}, nil
}

// GetPopularVideos 热门排行榜。
func (s *VideoServiceImpl) GetPopularVideos(
	ctx context.Context, req *video.GetPopularVideosRequest,
) (*video.GetPopularVideosResponse, error) {
	items, total, err := s.svc.Popular(ctx,
		int(derefInt64(req.PageNum)), int(derefInt64(req.PageSize)))
	if err != nil {
		return &video.GetPopularVideosResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &video.GetPopularVideosResponse{
		Base:  rpcbase.OK(),
		Items: pack.BuildVideos(items),
		Total: &total,
	}, nil
}

// SearchVideos 多条件搜索。
func (s *VideoServiceImpl) SearchVideos(
	ctx context.Context, req *video.SearchVideosRequest,
) (*video.SearchVideosResponse, error) {
	items, total, err := s.svc.Search(ctx, service.SearchParams{
		Keywords: derefString(req.Keywords),
		Username: derefString(req.Username),
		FromTime: millisToTime(req.FromMillis),
		ToTime:   millisToTime(req.ToMillis),
		PageNum:  int(derefInt64(req.PageNum)),
		PageSize: int(derefInt64(req.PageSize)),
	})
	if err != nil {
		return &video.SearchVideosResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &video.SearchVideosResponse{
		Base:  rpcbase.OK(),
		Items: pack.BuildVideos(items),
		Total: &total,
	}, nil
}

// VisitVideo 记录播放。
func (s *VideoServiceImpl) VisitVideo(
	ctx context.Context, req *video.VisitVideoRequest,
) (*video.VisitVideoResponse, error) {
	if err := s.svc.Visit(ctx, req.VideoId); err != nil {
		return &video.VisitVideoResponse{Base: rpcbase.Fail(err)}, nil
	}
	return &video.VisitVideoResponse{Base: rpcbase.OK()}, nil
}

// millisToTime 把毫秒时间戳转成时间；nil 或非正数表示不限制。
func millisToTime(v *int64) *time.Time {
	if v == nil || *v <= 0 {
		return nil
	}
	t := time.UnixMilli(*v)
	return &t
}

// derefInt64 解引用可选的整型字段。
func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// derefString 解引用可选的字符串字段。
func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
