package pack

import (
	hzmodel "gobili/api/model/model"
	"gobili/kitex_gen/rpcmodel"
)

// BuildVideo 把 RPC 返回的视频模型转成对外的 HTTP 模型。
func BuildVideo(v *rpcmodel.Video) *hzmodel.Video {
	if v == nil {
		return nil
	}
	return &hzmodel.Video{
		ID:           FormatID(v.Id),
		UserID:       FormatID(v.UserId),
		VideoURL:     ObjectURL(v.VideoKey),
		CoverURL:     ObjectURL(v.CoverKey),
		Title:        v.Title,
		Description:  v.Description,
		VisitCount:   v.VisitCount,
		CommentCount: v.CommentCount,
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
	}
}

// BuildVideos 批量转换，保证空列表返回空切片而不是 nil。
func BuildVideos(videos []*rpcmodel.Video) []*hzmodel.Video {
	result := make([]*hzmodel.Video, 0, len(videos))
	for _, v := range videos {
		result = append(result, BuildVideo(v))
	}
	return result
}
