package video

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// GetByID 按主键查视频。第二个返回值为 false 表示不存在（不是错误）。
func (d *DBVideo) GetByID(ctx context.Context, id int64) (bool, *model.Video, error) {
	v := new(model.Video)

	err := d.client.WithContext(ctx).Table(videoTable).Where("id = ?", id).First(v).Error
	if err == nil {
		return true, v, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil, nil
	}
	logger.Errorf("dal.video.GetByID: %v", err)
	return false, nil, errno.InternalError
}
