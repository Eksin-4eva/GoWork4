package video

import (
	"context"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// Count 返回未删除的视频总数。
//
// 热门榜走数据库回退路径时需要它给出 total，
// 否则只能拿「本页条数」冒充总数，客户端翻页会算错。
func (d *DBVideo) Count(ctx context.Context) (int64, error) {
	var total int64

	err := d.client.WithContext(ctx).Table(videoTable).Count(&total).Error
	if err != nil {
		logger.Errorf("dal.video.Count: %v", err)
		return 0, errno.InternalError
	}
	return total, nil
}
