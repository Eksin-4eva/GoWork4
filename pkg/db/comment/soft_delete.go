package comment

import (
	"context"

	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// SoftDelete 软删一条评论，并同步回滚相关计数。
//
// 调用方必须已经校验过「这条评论属于当前用户」——本方法只负责写。
// WHERE 带 deleted_at IS NULL，所以并发重复删除时只有第一次真正生效，
// 第二个返回值为 false 表示本次没删到（已经被删过了）。
func (d *DBComment) SoftDelete(ctx context.Context, id, videoID, parentID int64) (bool, error) {
	var affected bool

	err := d.client.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Table(commentTable).
			Where("id = ? AND deleted_at IS NULL", id).
			Update("deleted_at", gorm.Expr("NOW()"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		affected = true

		if err := tx.Table(constants.VideoTableName).
			Where("id = ? AND deleted_at IS NULL AND comment_count > 0", videoID).
			UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error; err != nil {
			return err
		}
		if parentID == 0 {
			return nil
		}
		return tx.Table(commentTable).
			Where("id = ? AND deleted_at IS NULL AND child_count > 0", parentID).
			UpdateColumn("child_count", gorm.Expr("child_count - 1")).Error
	})
	if err != nil {
		logger.Errorf("dal.comment.SoftDelete: %v", err)
		return false, errno.InternalError
	}
	return affected, nil
}
