package pack

import (
	"strconv"
	"strings"

	"gobili/config"
)

// FormatID 把 RPC 层的 i64 主键转成对外的字符串。
//
// 对外一律用字符串：int64 超过 2^53 后在 JavaScript 里会丢精度，
// 而雪花 ID 是 19 位。
func FormatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

// FormatIDPtr 处理可选的 i64。
func FormatIDPtr(id *int64) string {
	if id == nil {
		return ""
	}
	return FormatID(*id)
}

// ObjectURL 把对象存储的 key 拼成客户端可直接访问的完整地址。
//
// **这是 key → URL 的唯一转换点**，也是微服务拆分后网关的独有职责：
// 下游服务只传 key（它们不知道客户端从哪个域名访问），
// 由网关按当前部署的 PublicBaseURL 拼出地址。
//
// 好处是换存储或换 CDN 只改一行配置，不用刷任何数据。
func ObjectURL(key string) string {
	if key == "" {
		// 空 key 返回空串，而不是拼一个取不到的假地址。
		// 旧版就是拼了个必然 404 的封面地址，前端只能拿到坏图。
		return ""
	}
	base := strings.TrimSuffix(config.Get().MinIO.PublicBaseURL, "/")
	return base + "/" + strings.TrimPrefix(key, "/")
}
