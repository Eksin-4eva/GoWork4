// Package service 实现视频域的业务逻辑。
package service

import (
	"gobili/pkg/base"
	"gobili/pkg/cache"
	"gobili/pkg/db"
	"gobili/pkg/kafka"
	"gobili/pkg/storage"
	"gobili/pkg/utils"
)

// VideoService 是视频域的业务实现。
type VideoService struct {
	db      *db.Database
	cache   *cache.Cache
	storage *storage.Storage
	mq      *kafka.Broker
	sf      *utils.Snowflake
}

// New 构造视频域服务。依赖取自进程级 ClientSet。
func New() *VideoService {
	cs := base.GetClientSet()
	return &VideoService{
		db:      cs.DB,
		cache:   cs.Cache,
		storage: cs.Storage,
		mq:      cs.MQ,
		sf:      cs.SF,
	}
}
