package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	"gobili/api/router"
)

func main() {
	h := server.Default()

	router.Register(h)

	h.Spin()
}
