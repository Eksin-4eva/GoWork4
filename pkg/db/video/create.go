package video

import (
	"context"

	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Create 插入一条视频记录。
func (d *DBVideo) Create(ctx context.Context, v *model.Video) error {
	err := d.client.WithContext(ctx).Table(videoTable).Create(v).Error
	if err != nil {
		logger.Errorf("dal.video.Create: %v", err)
		return errno.InternalError
	}
	return nil
}
