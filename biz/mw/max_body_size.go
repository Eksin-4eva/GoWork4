package mw

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func MaxBodySize(maxBytes int64) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 检查Content-Length头
		contentLength := int64(c.Request.Header.ContentLength())
		if contentLength > 0 && contentLength > maxBytes {
			c.JSON(consts.StatusRequestEntityTooLarge, map[string]interface{}{
				"code":    consts.StatusRequestEntityTooLarge,
				"message": "request body too large",
			})
			c.Abort()
			return
		}
		// 对于分块传输编码，我们信任服务器级别的限制
		c.Next(ctx)
	}
}
