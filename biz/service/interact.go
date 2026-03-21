package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/BiliGO/biz/dal/model"
	"github.com/BiliGO/biz/dal/mysql"
	"github.com/BiliGO/biz/dal/query"
	"gorm.io/gorm"
)

func LikeVideo(ctx context.Context, userID, videoID int64, actionType int) error {
	q := query.Use(mysql.DB)
	fq := q.Favorite
	vq := q.Video

	existing, err := fq.WithContext(ctx).Where(fq.UserID.Eq(userID), fq.VideoID.Eq(videoID)).First()

	if actionType == 1 {
		if err == nil && existing != nil {
			return errors.New("already liked")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			return fmt.Errorf("db error: %w", err)
		}
		if err := fq.WithContext(ctx).Create(&model.Favorite{UserID: userID, VideoID: videoID}); err != nil {
			return err
		}
		_, err = vq.WithContext(ctx).Where(vq.ID.Eq(videoID)).UpdateSimple(vq.LikeCount.Add(1))
		return err
	}

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

func LikeComment(ctx context.Context, userID, commentID int64, actionType int) error {
	q := query.Use(mysql.DB)
	cq := q.Comment

	_, err := cq.WithContext(ctx).Where(cq.ID.Eq(commentID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("comment not found")
	}
	if err != nil {
		return err
	}

	var existingLike *model.CommentLike
	err = mysql.DB.WithContext(ctx).Where("user_id = ? AND comment_id = ?", userID, commentID).First(&existingLike).Error

	if actionType == 1 {
		if err == nil && existingLike != nil {
			return errors.New("already liked")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			return fmt.Errorf("db error: %w", err)
		}
		like := &model.CommentLike{UserID: userID, CommentID: commentID}
		if err := mysql.DB.WithContext(ctx).Create(like).Error; err != nil {
			return err
		}
		_, err = cq.WithContext(ctx).Where(cq.ID.Eq(commentID)).UpdateSimple(cq.LikeCount.Add(1))
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("not liked yet")
	}
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}
	if err := mysql.DB.WithContext(ctx).Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&model.CommentLike{}).Error; err != nil {
		return err
	}
	_, err = cq.WithContext(ctx).Where(cq.ID.Eq(commentID)).UpdateSimple(cq.LikeCount.Sub(1))
	return err
}

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
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	VideoID    string `json:"video_id"`
	ParentID   string `json:"parent_id"`
	LikeCount  int64  `json:"like_count"`
	ChildCount int64  `json:"child_count"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func commentToItem(c *model.Comment) CommentItem {
	return CommentItem{
		ID:         strconv.FormatInt(c.ID, 10),
		UserID:     strconv.FormatInt(c.UserID, 10),
		VideoID:    strconv.FormatInt(c.VideoID, 10),
		ParentID:   strconv.FormatInt(c.ParentID, 10),
		LikeCount:  c.LikeCount,
		ChildCount: c.ChildCount,
		Content:    c.Content,
		CreatedAt:  c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func PublishComment(ctx context.Context, userID int64, videoID int64, parentID int64, content string) (*CommentItem, error) {
	if content == "" {
		return nil, errors.New("content is required")
	}
	q := query.Use(mysql.DB)

	if parentID > 0 {
		parent, err := q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(parentID)).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("parent comment not found")
		}
		if err != nil {
			return nil, err
		}
		videoID = parent.VideoID
		_, err = q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(parentID)).UpdateSimple(q.Comment.ChildCount.Add(1))
		if err != nil {
			return nil, err
		}
	} else {
		_, err := q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("video not found")
		}
		if err != nil {
			return nil, err
		}
	}

	c := &model.Comment{
		VideoID:  videoID,
		UserID:   userID,
		Content:  content,
		ParentID: parentID,
	}
	if err := q.Comment.WithContext(ctx).Create(c); err != nil {
		return nil, err
	}

	if parentID == 0 {
		q.Video.WithContext(ctx).Where(q.Video.ID.Eq(videoID)).UpdateSimple(q.Video.CommentCount.Add(1))
	}

	item := commentToItem(c)
	return &item, nil
}

func GetCommentList(ctx context.Context, videoID int64, parentID int64, pageNum, pageSize int) ([]CommentItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)

	var cq query.ICommentDo
	if parentID > 0 {
		cq = q.Comment.WithContext(ctx).Where(q.Comment.ParentID.Eq(parentID)).Order(q.Comment.CreatedAt.Desc())
	} else {
		cq = q.Comment.WithContext(ctx).Where(q.Comment.VideoID.Eq(videoID), q.Comment.ParentID.Eq(0)).Order(q.Comment.CreatedAt.Desc())
	}

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

func DeleteComment(ctx context.Context, userID int64, commentID int64) error {
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

	// 递归删除子评论
	if err := deleteChildComments(ctx, q, commentID); err != nil {
		return err
	}

	if c.ParentID > 0 {
		q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(c.ParentID)).UpdateSimple(q.Comment.ChildCount.Sub(1))
	} else {
		q.Video.WithContext(ctx).Where(q.Video.ID.Eq(c.VideoID)).UpdateSimple(q.Video.CommentCount.Sub(1))
	}

	_, err = q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(commentID)).Delete()
	return err
}

// deleteChildComments 递归删除子评论
func deleteChildComments(ctx context.Context, q *query.Query, parentID int64) error {
	// 查询所有子评论
	childComments, err := q.Comment.WithContext(ctx).Where(q.Comment.ParentID.Eq(parentID)).Find()
	if err != nil {
		return err
	}

	// 递归删除每个子评论的子评论
	for _, child := range childComments {
		if err := deleteChildComments(ctx, q, child.ID); err != nil {
			return err
		}

		// 删除子评论本身
		_, err := q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(child.ID)).Delete()
		if err != nil {
			return err
		}

		// 更新父评论的子评论计数
		if child.ParentID > 0 {
			q.Comment.WithContext(ctx).Where(q.Comment.ID.Eq(child.ParentID)).UpdateSimple(q.Comment.ChildCount.Sub(1))
		} else {
			// 如果是顶级评论，更新视频的评论计数
			q.Video.WithContext(ctx).Where(q.Video.ID.Eq(child.VideoID)).UpdateSimple(q.Video.CommentCount.Sub(1))
		}
	}

	return nil
}
