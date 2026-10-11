// Package service 实现视频域的业务逻辑。
package service

import (
	"gobili/pkg/base"
	"gobili/pkg/cache"
	"gobili/pkg/db"
	"gobili/pkg/kafka"
	"gobili/pkg/utils"
)

// VideoService 是视频域的业务实现。
//
// 微服务化之后它**不再持有对象存储**：文件由网关上传递到 MinIO，
// 本服务只接收 object key；只有 worker 需要下载视频来抽帧。
type VideoService struct {
	db    *db.Database
	cache *cache.Cache
	mq    *kafka.Broker
	sf    *utils.Snowflake
}

// New 构造视频域服务。依赖取自进程级 ClientSet。
func New() *VideoService {
	cs := base.GetClientSet()
	return &VideoService{
		db:    cs.DB,
		cache: cs.Cache,
		mq:    cs.MQ,
		sf:    cs.SF,
	}
}
