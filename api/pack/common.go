package pack

import (
	"strconv"
	"strings"
	"time"

	"gobili/config"
)

// FormatID 把雪花 ID 转成字符串。
//
// 对外一律用字符串：int64 超过 2^53 后在 JavaScript 里会丢精度。
func FormatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

// FormatTime 把数据库时间转成对外字符串；零值返回空串，
// 避免把 "0001-01-01 00:00:00" 这种零值透给客户端。
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// ObjectURL 把 MinIO 的 object key 拼成可直接访问的完整地址。
//
// 库里存的是 key 而不是 URL，所以换存储或换域名只需要改配置，不用刷数据。
func ObjectURL(key string) string {
	if key == "" {
		return ""
	}
	base := strings.TrimSuffix(config.Get().MinIO.PublicBaseURL, "/")
	return base + "/" + strings.TrimPrefix(key, "/")
}
