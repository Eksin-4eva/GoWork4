// Package service 实现用户域的业务逻辑。
package service

import (
	"github.com/minio/minio-go/v7"

	"gobili/pkg/base"
	"gobili/pkg/db"
	"gobili/pkg/utils"
)

// UserService 是用户域的业务实现。
type UserService struct {
	db      *db.Database
	storage *minio.Client
	sf      *utils.Snowflake
}

// New 构造用户域服务。依赖取自进程级 ClientSet，由 main 在启动时初始化。
func New() *UserService {
	cs := base.GetClientSet()
	return &UserService{db: cs.DB, storage: cs.Storage, sf: cs.SF}
}
