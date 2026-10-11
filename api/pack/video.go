package pack

import (
	hzmodel "gobili/api/model/model"
	dbmodel "gobili/pkg/db/model"
)

// BuildVideo 把数据库模型转成对外的视频模型。
//
// VideoKey / CoverKey 是对象存储的 key，这里才拼成可访问的 URL。
// CoverKey 为空表示封面尚未生成（Phase 8 由 worker 异步抽帧补上），
// 此时返回空字符串，而不是旧版那种拼出来必然 404 的假地址。
func BuildVideo(v *dbmodel.Video) *hzmodel.Video {
	if v == nil {
		return nil
	}
	return &hzmodel.Video{
		ID:           FormatID(v.ID),
		UserID:       FormatID(v.UserID),
		VideoURL:     ObjectURL(v.VideoKey),
		CoverURL:     ObjectURL(v.CoverKey),
		Title:        v.Title,
		Description:  v.Description,
		VisitCount:   v.VisitCount,
		CommentCount: v.CommentCount,
		CreatedAt:    FormatTime(v.CreatedAt),
		UpdatedAt:    FormatTime(v.UpdatedAt),
	}
}

// BuildVideos 批量转换，保证空列表返回 [] 而不是 null。
func BuildVideos(videos []*dbmodel.Video) []*hzmodel.Video {
	result := make([]*hzmodel.Video, 0, len(videos))
	for _, v := range videos {
		result = append(result, BuildVideo(v))
	}
	return result
}
