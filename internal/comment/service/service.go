// Package service 实现评论域的业务逻辑。
package service

import (
	"gobili/pkg/base"
	"gobili/pkg/db"
	"gobili/pkg/utils"
)

// CommentService 是评论域的业务实现。
type CommentService struct {
	db *db.Database
	sf *utils.Snowflake
}

// New 构造评论域服务。依赖取自进程级 ClientSet。
func New() *CommentService {
	cs := base.GetClientSet()
	return &CommentService{db: cs.DB, sf: cs.SF}
}
