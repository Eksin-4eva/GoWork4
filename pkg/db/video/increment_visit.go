package video

import (
	"context"

	"gorm.io/gorm"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// IncrementVisit 给播放数加一。
//
// 第二个返回值为 false 表示该视频不存在（或已软删），
// 用于替代旧版"先查再更新"的存在性校验。
//
// 用 UpdateColumn 而非 Update：计数器不该顺带改写 updated_at，
// 否则每次播放都会把整个索引行重写一遍。
func (d *DBVideo) IncrementVisit(ctx context.Context, id int64) (bool, error) {
	res := d.client.WithContext(ctx).Table(videoTable).
		Where("id = ? AND "+notDeleted, id).
		UpdateColumn("visit_count", gorm.Expr("visit_count + 1"))
	if res.Error != nil {
		logger.Errorf("dal.video.IncrementVisit: %v", res.Error)
		return false, errno.InternalError
	}
	return res.RowsAffected > 0, nil
}
