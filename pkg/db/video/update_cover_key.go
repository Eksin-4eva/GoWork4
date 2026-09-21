package video

import (
	"context"

	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// SetCoverKey 回写封面 key。
//
// WHERE 里带上「cover_key 为空」的判断作为幂等护栏：Kafka 至少一次投递会让 worker
// 重复处理同一条消息，把判断和写入合成一条 SQL 才能避免"先查后写"的竞态。
// 第二个返回值为 false 表示该视频已经有封面（或不存在），本次写入被跳过。
func (d *DBVideo) SetCoverKey(ctx context.Context, id int64, key string) (bool, error) {
	res := d.client.WithContext(ctx).Table(videoTable).
		Where("id = ? AND "+notDeleted+" AND cover_key = ''", id).
		Update("cover_key", key)
	if res.Error != nil {
		logger.Errorf("dal.video.SetCoverKey: %v", res.Error)
		return false, errno.InternalError
	}
	return res.RowsAffected > 0, nil
}
