// Package pack 负责「数据库模型 → RPC 模型」的转换。
package pack

import (
	"time"

	"gobili/kitex_gen/rpcmodel"
	dbmodel "gobili/pkg/db/model"
)

// timeLayout 是 RPC 层传递时间的统一格式。
const timeLayout = "2006-01-02 15:04:05"

// BuildVideo 把数据库模型转成 RPC 模型。
//
// 注意传的是 video_key / cover_key，不是 URL——
// 拼 URL 是网关的职责（取决于客户端从哪个域名访问）。
func BuildVideo(v *dbmodel.Video) *rpcmodel.Video {
	if v == nil {
		return nil
	}
	return &rpcmodel.Video{
		Id:           v.ID,
		UserId:       v.UserID,
		VideoKey:     v.VideoKey,
		CoverKey:     v.CoverKey,
		Title:        v.Title,
		Description:  v.Description,
		VisitCount:   v.VisitCount,
		CommentCount: v.CommentCount,
		CreatedAt:    formatTime(v.CreatedAt),
		UpdatedAt:    formatTime(v.UpdatedAt),
	}
}

// BuildVideos 批量转换，保证空列表返回空切片而不是 nil。
func BuildVideos(videos []*dbmodel.Video) []*rpcmodel.Video {
	result := make([]*rpcmodel.Video, 0, len(videos))
	for _, v := range videos {
		result = append(result, BuildVideo(v))
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
