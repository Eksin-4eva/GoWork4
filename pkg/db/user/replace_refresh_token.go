package user

import (
	"context"

	"gorm.io/gorm"

	"gobili/pkg/constants"
	"gobili/pkg/db/model"
	"gobili/pkg/errno"
	"gobili/pkg/logger"
)

// ReplaceRefreshToken 在一个事务里「吊销旧刷新令牌 + 落库新刷新令牌」。
//
// 返回 false 表示旧令牌不可用：不存在、已吊销、已过期，或者不属于这个用户。
//
// **并发安全靠的是 SQL，不是锁**：「仍然有效」这几个条件直接写进 UPDATE 的
// WHERE 里，所以并发的两次换发只有一次能把 RowsAffected 变成 1，
// 另一次会拿到 0 —— 不会凭空产生两个有效令牌。
// 这与 SetCoverKey 用的是同一个套路：能用原子 SQL 解决的就不加锁。
//
// 注意 WHERE 里那个 deleted_at IS NULL：写操作用的是 .Table(表名) 而不是
// .Model(结构体)，GORM 拿不到 schema，不会自动补软删条件，必须自己带上
// （与 pkg/db/video/video.go 里 notDeleted 是同一个坑）。
func (d *DBUser) ReplaceRefreshToken(
	ctx context.Context, old string, userID int64, next *model.RefreshToken,
) (bool, error) {
	swapped := false

	err := d.client.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Table(constants.RefreshTokenTableName).
			Where("token = ? AND user_id = ? AND revoked = 0 AND expires_at > NOW() AND "+
				"deleted_at IS NULL", old, userID).
			Update("revoked", 1)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// 旧令牌不可用：什么都没改到，直接结束（也不需要回滚）。
			return nil
		}

		if err := tx.Table(constants.RefreshTokenTableName).Create(next).Error; err != nil {
			return err
		}
		swapped = true
		return nil
	})
	if err != nil {
		logger.Errorf("dal.user.ReplaceRefreshToken: %v", err)
		return false, errno.InternalError
	}
	return swapped, nil
}
