package comment

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// GetByID 按主键查评论。第二个返回值为 false 表示不存在（不是错误）。
func (d *DBComment) GetByID(ctx context.Context, id int64) (bool, *model.Comment, error) {
	c := new(model.Comment)

	err := d.client.WithContext(ctx).Table(commentTable).Where("id = ?", id).First(c).Error
	if err == nil {
		return true, c, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil, nil
	}
	logger.Errorf("dal.comment.GetByID: %v", err)
	return false, nil, errno.InternalError
}
