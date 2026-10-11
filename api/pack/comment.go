package pack

import (
	hzmodel "gobili/api/model/model"
	"gobili/kitex_gen/rpcmodel"
)

// BuildComment 把 RPC 返回的评论模型转成对外的 HTTP 模型。
func BuildComment(c *rpcmodel.Comment) *hzmodel.Comment {
	if c == nil {
		return nil
	}
	return &hzmodel.Comment{
		ID:         FormatID(c.Id),
		UserID:     FormatID(c.UserId),
		VideoID:    FormatID(c.VideoId),
		ParentID:   FormatID(c.ParentId),
		ChildCount: c.ChildCount,
		Content:    c.Content,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// BuildComments 批量转换，保证空列表返回空切片而不是 nil。
func BuildComments(comments []*rpcmodel.Comment) []*hzmodel.Comment {
	result := make([]*hzmodel.Comment, 0, len(comments))
	for _, c := range comments {
		result = append(result, BuildComment(c))
	}
	return result
}
