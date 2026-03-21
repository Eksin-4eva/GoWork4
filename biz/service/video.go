package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BiliGO/biz/dal/model"
	"github.com/BiliGO/biz/dal/mysql"
	"github.com/BiliGO/biz/dal/query"
	"github.com/BiliGO/biz/dal/redis"
	"gorm.io/gorm"
)

type VideoItem struct {
	ID           string
	UserID       string
	Title        string
	Description  string
	VideoURL     string
	CoverURL     string
	VisitCount   int64
	LikeCount    int64
	CommentCount int64
	CreatedAt    string
	UpdatedAt    string
}

func getMinioURL(storedURL string) string {
	if storedURL == "" {
		return ""
	}
	if strings.HasPrefix(storedURL, "http://") || strings.HasPrefix(storedURL, "https://") {
		return storedURL
	}
	// 优先使用外部访问地址（给浏览器用）
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "biligo"
	}
	return fmt.Sprintf("http://localhost:9000/%s/%s", bucket, storedURL)
}

func videoToItem(v *model.Video) VideoItem {
	return VideoItem{
		ID:           strconv.FormatInt(v.ID, 10),
		UserID:       strconv.FormatInt(v.UserID, 10),
		Title:        v.Title,
		Description:  v.Description,
		VideoURL:     getMinioURL(v.VideoURL),
		CoverURL:     getMinioURL(v.CoverURL),
		VisitCount:   v.VisitCount,
		LikeCount:    v.LikeCount,
		CommentCount: v.CommentCount,
		CreatedAt:    v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    v.UpdatedAt.Format(time.RFC3339),
	}
}

// PublishVideo 投稿视频
func PublishVideo(ctx context.Context, userID int64, title, description, videoURL, coverURL string) (*VideoItem, error) {
	if title == "" || videoURL == "" {
		return nil, errors.New("title and video_url are required")
	}
	q := query.Use(mysql.DB)
	v := &model.Video{
		UserID:      userID,
		Title:       title,
		Description: description,
		VideoURL:    videoURL,
		CoverURL:    coverURL,
	}
	if err := q.Video.WithContext(ctx).Create(v); err != nil {
		return nil, fmt.Errorf("create video failed: %w", err)
	}
	item := videoToItem(v)
	return &item, nil
}

// GetVideoList 获取用户发布列表（分页）
func GetVideoList(ctx context.Context, userID int64, pageNum, pageSize int) ([]VideoItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)
	vq := q.Video.WithContext(ctx).Where(q.Video.UserID.Eq(userID)).Order(q.Video.CreatedAt.Desc())
	total, err := vq.Count()
	if err != nil {
		return nil, 0, err
	}
	videos, err := vq.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		items[i] = videoToItem(v)
	}
	return items, total, nil
}

// GetPopularVideos 热门排行榜（按播放量降序）
func GetPopularVideos(ctx context.Context, pageNum, pageSize int) ([]VideoItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)
	vq := q.Video.WithContext(ctx).Order(q.Video.VisitCount.Desc())
	total, err := vq.Count()
	if err != nil {
		return nil, 0, err
	}
	videos, err := vq.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}

	// 获取视频ID列表
	videoIDs := make([]int64, len(videos))
	for i, v := range videos {
		videoIDs[i] = v.ID
	}

	// 从 Redis 批量获取播放量
	redisCounts, err := redis.BatchGetVideoVisitCounts(ctx, videoIDs)
	if err != nil {
		// Redis 失败时使用数据库中的播放量
		items := make([]VideoItem, len(videos))
		for i, v := range videos {
			items[i] = videoToItem(v)
		}
		return items, total, nil
	}

	// 使用 Redis 中的播放量
	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		item := videoToItem(v)
		if count, ok := redisCounts[v.ID]; ok {
			item.VisitCount = count
		}
		items[i] = item
	}

	return items, total, nil
}

// SearchVideo 搜索视频（title + description 模糊匹配，分页）
func SearchVideo(ctx context.Context, keyword string, pageNum, pageSize int) ([]VideoItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)
	like := "%" + keyword + "%"
	vq := q.Video.WithContext(ctx).Where(
		q.Video.Title.Like(like),
	).Or(
		q.Video.Description.Like(like),
	).Order(q.Video.CreatedAt.Desc())

	total, err := vq.Count()
	if err != nil {
		return nil, 0, err
	}
	videos, err := vq.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		items[i] = videoToItem(v)
	}
	return items, total, nil
}

// GetVideoByID 按 ID 查视频
func GetVideoByID(ctx context.Context, videoID int64) (*model.Video, error) {
	q := query.Use(mysql.DB)
	v, err := q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("video not found")
	}
	return v, err
}

// GetVideoDetail 获取视频详情并增加访问次数
func GetVideoDetail(ctx context.Context, videoID int64) (*VideoItem, error) {
	q := query.Use(mysql.DB)

	// 查询视频
	v, err := q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("video not found")
	}
	if err != nil {
		return nil, err
	}

	// 使用 Redis 增加访问次数
	visitCount, err := redis.IncrementVideoVisitCount(ctx, videoID)
	if err != nil {
		// Redis 失败时回退到数据库
		_, err = q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).UpdateSimple(q.Video.VisitCount.Add(1))
		if err != nil {
			return nil, err
		}
		v.VisitCount++
	} else {
		v.VisitCount = visitCount
	}

	item := videoToItem(v)
	return &item, nil
}
