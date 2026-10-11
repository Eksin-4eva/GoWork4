// Package service 实现用户域的业务逻辑。
package service

import (
	"gobili/pkg/base"
	"gobili/pkg/db"
	"gobili/pkg/utils"
)

// UserService 是用户域的业务实现。
//
// 微服务化之后它**不再持有对象存储**：文件由网关上传递到 MinIO，
// 本服务只接收 object key。
type UserService struct {
	db *db.Database
	sf *utils.Snowflake
}

// New 构造用户域服务。依赖取自进程级 ClientSet，由 main 在启动时初始化。
func New() *UserService {
	cs := base.GetClientSet()
	return &UserService{db: cs.DB, sf: cs.SF}
}
