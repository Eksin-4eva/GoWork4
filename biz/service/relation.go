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

type UserItem struct {
	ID             string
	Username       string
	AvatarURL      string
	FollowerCount  int64
	FollowingCount int64
	CreatedAt      string
}

func userToItem(u *model.User) UserItem {
	return UserItem{
		ID:             strconv.FormatInt(u.ID, 10),
		Username:       u.Username,
		AvatarURL:      u.AvatarURL,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
		CreatedAt:      u.CreatedAt.Format(time.RFC3339),
	}
}

// RelationAction 关注/取消关注 (actionType: 1=关注, 2=取消)
func RelationAction(ctx context.Context, userID, toUserID int64, actionType int) error {
	if userID == toUserID {
		return errors.New("cannot follow yourself")
	}
	q := query.Use(mysql.DB)
	rq := q.Relation
	uq := q.User

	existing, err := rq.WithContext(ctx).Where(rq.UserID.Eq(userID), rq.ToUserID.Eq(toUserID)).First()

	if actionType == 1 {
		// 关注
		if err == nil && existing.Status == 1 {
			return errors.New("already following")
		}
		if err == nil {
			// 记录存在但 status=0，更新为 1
			_, err = rq.WithContext(ctx).Where(rq.UserID.Eq(userID), rq.ToUserID.Eq(toUserID)).
				UpdateSimple(rq.Status.Value(1))
			if err != nil {
				return fmt.Errorf("update relation failed: %w", err)
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := rq.WithContext(ctx).Create(&model.Relation{UserID: userID, ToUserID: toUserID, Status: 1}); err != nil {
				return fmt.Errorf("create relation failed: %w", err)
			}
		} else {
			return fmt.Errorf("db error: %w", err)
		}
		// 更新计数
		uq.WithContext(ctx).Where(uq.ID.Eq(userID)).UpdateSimple(uq.FollowingCount.Add(1))
		uq.WithContext(ctx).Where(uq.ID.Eq(toUserID)).UpdateSimple(uq.FollowerCount.Add(1))
		return nil
	}

	// 取消关注
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && existing.Status == 0) {
		return errors.New("not following")
	}
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}
	_, err = rq.WithContext(ctx).Where(rq.UserID.Eq(userID), rq.ToUserID.Eq(toUserID)).
		UpdateSimple(rq.Status.Value(0))
	if err != nil {
		return err
	}
	uq.WithContext(ctx).Where(uq.ID.Eq(userID)).UpdateSimple(uq.FollowingCount.Sub(1))
	uq.WithContext(ctx).Where(uq.ID.Eq(toUserID)).UpdateSimple(uq.FollowerCount.Sub(1))
	return nil
}

// GetFollowingList 获取关注列表
func GetFollowingList(ctx context.Context, userID int64, pageNum, pageSize int) ([]UserItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)
	rq := q.Relation.WithContext(ctx).Where(q.Relation.UserID.Eq(userID), q.Relation.Status.Eq(1))
	total, err := rq.Count()
	if err != nil {
		return nil, 0, err
	}
	relations, err := rq.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	return fetchUserItems(ctx, relations, func(r *model.Relation) int64 { return r.ToUserID }), total, nil
}

// GetFollowerList 获取粉丝列表
func GetFollowerList(ctx context.Context, userID int64, pageNum, pageSize int) ([]UserItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)
	rq := q.Relation.WithContext(ctx).Where(q.Relation.ToUserID.Eq(userID), q.Relation.Status.Eq(1))
	total, err := rq.Count()
	if err != nil {
		return nil, 0, err
	}
	relations, err := rq.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, err
	}
	return fetchUserItems(ctx, relations, func(r *model.Relation) int64 { return r.UserID }), total, nil
}

// GetFriendsList 获取互相关注的好友列表
func GetFriendsList(ctx context.Context, userID int64, pageNum, pageSize int) ([]UserItem, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	q := query.Use(mysql.DB)

	// 我关注的人
	following, err := q.Relation.WithContext(ctx).
		Where(q.Relation.UserID.Eq(userID), q.Relation.Status.Eq(1)).Find()
	if err != nil {
		return nil, 0, err
	}
	followingIDs := make([]int64, len(following))
	for i, r := range following {
		followingIDs[i] = r.ToUserID
	}
	if len(followingIDs) == 0 {
		return []UserItem{}, 0, nil
	}

	// 在我关注的人中，找也关注了我的
	mutuals, err := q.Relation.WithContext(ctx).
		Where(q.Relation.ToUserID.Eq(userID), q.Relation.Status.Eq(1),
			q.Relation.UserID.In(followingIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(mutuals))
	start := (pageNum - 1) * pageSize
	if start >= int(total) {
		return []UserItem{}, total, nil
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	paged := mutuals[start:end]
	return fetchUserItems(ctx, paged, func(r *model.Relation) int64 { return r.UserID }), total, nil
}

func fetchUserItems(ctx context.Context, relations []*model.Relation, idFn func(*model.Relation) int64) []UserItem {
	if len(relations) == 0 {
		return []UserItem{}
	}
	ids := make([]int64, len(relations))
	for i, r := range relations {
		ids[i] = idFn(r)
	}
	q := query.Use(mysql.DB)
	users, err := q.User.WithContext(ctx).Where(q.User.ID.In(ids...)).Find()
	if err != nil {
		return []UserItem{}
	}
	items := make([]UserItem, len(users))
	for i, u := range users {
		items[i] = userToItem(u)
	}
	return items
}
