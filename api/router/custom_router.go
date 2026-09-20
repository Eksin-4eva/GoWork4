package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"gobili/api/handler"
)

// customizedRegister registers customized routers.
func customizedRegister(r *server.Hertz) {
	r.GET("/ping", handler.Ping)

	// your code ...
}
