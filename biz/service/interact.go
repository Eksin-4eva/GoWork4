package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/BiliGO/biz/dal/model"
	"github.com/BiliGO/biz/dal/mysql"
	"github.com/BiliGO/biz/dal/query"
	"gorm.io/gorm"
)

// LikeAction 点赞/取消点赞 (actionType: 1=点赞, 2=取消)
func LikeAction(ctx context.Context, userID, videoID int64, actionType int) error {
	q := query.Use(mysql.DB)
	fq := q.Favorite
	vq := q.Video

	existing, err := fq.WithContext(ctx).Where(fq.UserID.Eq(userID), fq.VideoID.Eq(videoID)).First()

	if actionType == 1 {
		// 点赞
		if err == nil && existing != nil {
			return errors.New("already liked")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			return fmt.Errorf("db error: %w", err)
		}
		if err := fq.WithContext(ctx).Create(&model.Favorite{UserID: userID, VideoID: videoID}); err != nil {
			return err
		}
		// 更新视频点赞数
		_, err = vq.WithContext(ctx).Where(vq.ID.Eq(videoID)).UpdateSimple(vq.LikeCount.Add(1))
		return err
	}

	// 取消点赞
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("not liked yet")
	}
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}
	if _, err := fq.WithContext(ctx).Where(fq.UserID.Eq(userID), fq.VideoID.Eq(videoID)).Delete(); err != nil {
		return err
	}
	_, err = vq.WithContext(ctx).Where(vq.ID.Eq(videoID)).UpdateSimple(vq.LikeCount.Sub(1))
	return err
}

type LikeItem struct {
	VideoID string
	Title   string
}

// GetLikeList 获取用户点赞列表
func GetLikeList(ctx context.Context, userID int64, pageNum, pageSize int) ([]VideoItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)
	fq := q.Favorite.WithContext(ctx).Where(q.Favorite.UserID.Eq(userID))
	total, err := fq.Count()
	if err != nil {
		return nil, 0, err
	}
	favs, err := fq.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	videoIDs := make([]int64, len(favs))
	for i, f := range favs {
		videoIDs[i] = f.VideoID
	}
	if len(videoIDs) == 0 {
		return []VideoItem{}, 0, nil
	}
	videos, err := q.Video.WithContext(ctx).Where(q.Video.ID.In(videoIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}
	items := make([]VideoItem, len(videos))
	for i, v := range videos {
		items[i] = videoToItem(v)
	}
	return items, total, nil
}

type CommentItem struct {
	ID        string
	VideoID   string
	UserID    string
	Content   string
	ParentID  string
	CreatedAt string
}

func commentToItem(c *model.Comment) CommentItem {
	return CommentItem{
		ID:        strconv.FormatInt(c.ID, 10),
		VideoID:   strconv.FormatInt(c.VideoID, 10),
		UserID:    strconv.FormatInt(c.UserID, 10),
		Content:   c.Content,
		ParentID:  strconv.FormatInt(c.ParentID, 10),
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

// PublishComment 发布评论
func PublishComment(ctx context.Context, userID, videoID int64, content string) (*CommentItem, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}
	q := query.Use(mysql.DB)
	// 检查视频存在
	_, err := q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("video not found")
	}
	if err != nil {
		return nil, err
	}
	c := &model.Comment{VideoID: videoID, UserID: userID, Content: content}
	if err := q.Comment.WithContext(ctx).Create(c); err != nil {
		return nil, err
	}
	// 更新评论数
	q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).UpdateSimple(q.Video.CommentCount.Add(1))
	item := commentToItem(c)
	return &item, nil
}

// GetCommentList 获取视频评论列表（分页）
func GetCommentList(ctx context.Context, videoID int64, pageNum, pageSize int) ([]CommentItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)
	cq := q.Comment.WithContext(ctx).Where(q.Comment.VideoID.Eq(videoID), q.Comment.ParentID.Eq(0)).Order(q.Comment.CreatedAt.Desc())
	total, err := cq.Count()
	if err != nil {
		return nil, 0, err
	}
	comments, err := cq.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	items := make([]CommentItem, len(comments))
	for i, c := range comments {
		items[i] = commentToItem(c)
	}
	return items, total, nil
}

// DeleteComment 删除评论（只能删自己的）
func DeleteComment(ctx context.Context, userID, commentID int64) error {
	q := query.Use(mysql.DB)
	c, err := q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(commentID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("comment not found")
	}
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return errors.New("cannot delete other's comment")
	}
	_, err = q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(commentID)).Delete()
	if err != nil {
		return err
	}
	// 更新评论数
	q.Video.WithContext(ctx).Where(q.Video.ID.Eq(c.VideoID)).UpdateSimple(q.Video.CommentCount.Sub(1))
	return nil
}
