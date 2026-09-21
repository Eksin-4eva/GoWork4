package video

import (
	"context"
	"time"

	"gorm.io/gorm"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
	"gobili/pkg/utils"
)

// SearchParams 是视频搜索的过滤条件，零值字段表示不加该条件。
type SearchParams struct {
	Keywords string
	Username string
	FromTime *time.Time
	ToTime   *time.Time
	PageNum  int
	PageSize int
}

// Search 按关键词 / 用户名 / 时间区间组合搜索视频，返回列表与总数。
//
// 注意 Count 与 Find 各自构造一次查询：GORM 的链式调用会累积条件，
// 复用同一个 *gorm.DB 做 Count 再 Find 容易把 count 的 SELECT 带进结果集。
func (d *DBVideo) Search(ctx context.Context, p SearchParams) ([]*model.Video, int64, error) {
	pageNum, pageSize := utils.NormalizePage(p.PageNum, p.PageSize)

	build := func() *gorm.DB {
		tx := d.client.WithContext(ctx).Table(videoTable)
		if p.Username != "" {
			tx = tx.Joins(searchJoin).
				Where(userTable+".username LIKE ?", "%"+p.Username+"%")
		}
		if p.Keywords != "" {
			kw := "%" + p.Keywords + "%"
			tx = tx.Where(videoTable+".title LIKE ? OR "+videoTable+".description LIKE ?", kw, kw)
		}
		if p.FromTime != nil {
			tx = tx.Where(videoTable+".created_at >= ?", *p.FromTime)
		}
		if p.ToTime != nil {
			tx = tx.Where(videoTable+".created_at <= ?", *p.ToTime)
		}
		return tx
	}

	var total int64
	if err := build().Count(&total).Error; err != nil {
		logger.Errorf("dal.video.Search count: %v", err)
		return nil, 0, errno.InternalError
	}

	items := make([]*model.Video, 0, pageSize)
	err := build().
		Order(orderLatest).
		Limit(pageSize).
		Offset(utils.Offset(pageNum, pageSize)).
		Find(&items).Error
	if err != nil {
		logger.Errorf("dal.video.Search: %v", err)
		return nil, 0, errno.InternalError
	}
	return items, total, nil
}
