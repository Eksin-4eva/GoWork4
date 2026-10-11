// Package service 实现视频域的业务逻辑。
package service

import (
	"github.com/minio/minio-go/v7"

	"gobili/pkg/base"
	"gobili/pkg/cache"
	"gobili/pkg/db"
	"gobili/pkg/utils"
)

// VideoService 是视频域的业务实现。
type VideoService struct {
	db      *db.Database
	cache   *cache.Cache
	storage *minio.Client
	sf      *utils.Snowflake
}

// New 构造视频域服务。依赖取自进程级 ClientSet。
func New() *VideoService {
	cs := base.GetClientSet()
	return &VideoService{db: cs.DB, cache: cs.Cache, storage: cs.Storage, sf: cs.SF}
}
