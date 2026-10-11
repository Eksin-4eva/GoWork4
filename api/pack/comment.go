package pack

import (
	hzmodel "gobili/api/model/model"
	dbmodel "gobili/pkg/db/model"
)

// BuildComment 把数据库模型转成对外的评论模型。
func BuildComment(c *dbmodel.Comment) *hzmodel.Comment {
	if c == nil {
		return nil
	}
	return &hzmodel.Comment{
		ID:         FormatID(c.ID),
		UserID:     FormatID(c.UserID),
		VideoID:    FormatID(c.VideoID),
		ParentID:   FormatID(c.ParentID),
		ChildCount: c.ChildCount,
		Content:    c.Content,
		CreatedAt:  FormatTime(c.CreatedAt),
		UpdatedAt:  FormatTime(c.UpdatedAt),
	}
}

// BuildComments 批量转换，保证空列表返回 [] 而不是 null。
func BuildComments(comments []*dbmodel.Comment) []*hzmodel.Comment {
	result := make([]*hzmodel.Comment, 0, len(comments))
	for _, c := range comments {
		result = append(result, BuildComment(c))
	}
	return result
}
