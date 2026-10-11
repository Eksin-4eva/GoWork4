// Package pack 负责「数据库模型 → RPC 模型」的转换。
package pack

import (
	"time"

	"gobili/kitex_gen/rpcmodel"
	dbmodel "gobili/pkg/db/model"
)

// timeLayout 是 RPC 层传递时间的统一格式。
const timeLayout = "2006-01-02 15:04:05"

// BuildComment 把数据库模型转成 RPC 模型。
func BuildComment(c *dbmodel.Comment) *rpcmodel.Comment {
	if c == nil {
		return nil
	}
	return &rpcmodel.Comment{
		Id:         c.ID,
		UserId:     c.UserID,
		VideoId:    c.VideoID,
		ParentId:   c.ParentID,
		ChildCount: c.ChildCount,
		Content:    c.Content,
		CreatedAt:  formatTime(c.CreatedAt),
		UpdatedAt:  formatTime(c.UpdatedAt),
	}
}

// BuildComments 批量转换，保证空列表返回空切片而不是 nil。
func BuildComments(comments []*dbmodel.Comment) []*rpcmodel.Comment {
	result := make([]*rpcmodel.Comment, 0, len(comments))
	for _, c := range comments {
		result = append(result, BuildComment(c))
	}
	return result
}

// formatTime 零值返回空串。
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(timeLayout)
}
