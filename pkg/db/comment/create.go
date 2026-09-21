package comment

import (
	"context"

	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Create 插入一条评论，并同步维护父评论的 child_count。
//
// 三条写放在同一事务里：评论本身成功但计数没跟上，
// 就会出现「点进去有回复、列表上却显示 0 条回复」这种不一致。
func (d *DBComment) Create(ctx context.Context, c *model.Comment) error {
	err := d.client.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(commentTable).Create(c).Error; err != nil {
			return err
		}
		// 视频的总评论数（含回复）
		if err := tx.Table(constants.VideoTableName).
			Where("id = ? AND deleted_at IS NULL", c.VideoID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
			return err
		}
		if c.ParentID == 0 {
			return nil
		}
		// 父评论的直接回复数
		return tx.Table(commentTable).
			Where("id = ? AND "+notDeleted, c.ParentID).
			UpdateColumn("child_count", gorm.Expr("child_count + 1")).Error
	})
	if err != nil {
		logger.Errorf("dal.comment.Create: %v", err)
		return errno.InternalError
	}
	return nil
}
