package api

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"gobili/api/mw"
	"gobili/pkg/base"
	"gobili/pkg/errno"
	"gobili/pkg/utils"
)

// currentUserID 取出鉴权中间件写入的当前用户 ID。
//
// 能走到 handler 说明中间件已经放行，正常不会取不到。
func currentUserID(c *app.RequestContext) int64 {
	return mw.CurrentUserID(c)
}

// parseID 把对外暴露的字符串 ID 解析成内部的 int64。
func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errno.ParamError.WithMessage("ID 格式不正确")
	}
	return id, nil
}

// parseOptionalID 解析可选 ID，空字符串表示「不指定」。
func parseOptionalID(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	return parseID(raw)
}

// derefString 解引用可选的字符串字段。
func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// derefInt 解引用可选的整型字段，nil 返回 0。
func derefInt(v *int64) int {
	if v == nil {
		return 0
	}
	return int(*v)
}

// millisToTime 把毫秒时间戳转成时间；nil 或非正数表示不限制。
func millisToTime(v *int64) *time.Time {
	if v == nil || *v <= 0 {
		return nil
	}
	t := time.UnixMilli(*v)
	return &t
}

// millisStringToTime 解析字符串形式的毫秒时间戳（feed 的 latest_time 是字符串）。
func millisStringToTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	ms, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || ms <= 0 {
		return nil, errno.ParamError.WithMessage("latest_time 需要是毫秒时间戳")
	}
	t := time.UnixMilli(ms)
	return &t, nil
}

// openUpload 打开 multipart 表单里的文件并包装成 utils.Upload。
// 返回的 release 必须由调用方 defer 执行。
func openUpload(c *app.RequestContext, field string) (upload utils.Upload, release func(), err error) {
	header, err := c.FormFile(field)
	if err != nil {
		return utils.Upload{}, nil, errno.ParamFile.WithMessage("缺少文件字段 " + field)
	}

	file, err := header.Open()
	if err != nil {
		return utils.Upload{}, nil, errno.ParamFile.WithMessage("无法读取上传的文件")
	}

	return utils.Upload{Reader: file, Size: header.Size, Filename: header.Filename},
		func() { _ = file.Close() }, nil
}

// storeUpload 把 multipart 文件**直接上传到对象存储**，返回 object key。
//
// 为什么由网关做这件事、而不是把文件字节传给下游服务：
// 走 RPC 传字节意味着几十上百 MB 的数据要在内存里序列化一次、反序列化一次，
// 而视频上限是 100MB。改成「网关传存储 + RPC 传 key」之后，
// 文件流只经过一次网络，下游服务完全不接触字节。
//
// objectID 只用于生成 key（如 video/123.mp4），与业务 ID 解耦。
func storeUpload(
	ctx context.Context, c *app.RequestContext,
	field, prefix string, objectID int64,
	allowed map[string]string, label string,
) (string, error) {
	upload, release, err := openUpload(c, field)
	if err != nil {
		return "", err
	}
	defer release()

	return base.GetClientSet().Storage.Put(ctx, prefix, objectID, allowed, upload, label)
}

// nextObjectID 生成一个用于对象 key 的雪花 ID。
func nextObjectID() (int64, error) {
	id, err := base.GetClientSet().SF.NextVal()
	if err != nil {
		return 0, errno.InternalError
	}
	return id, nil
}
