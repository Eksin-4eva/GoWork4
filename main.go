package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BiliGO/biz/dal/mysql"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/cors"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env（生产环境可直接注入环境变量，不依赖此文件）
	_ = godotenv.Load()

	if err := mysql.Init(&mysql.Config{
		DSN:             mustEnv("MYSQL_DSN"),
		MaxOpenConns:    envInt("MYSQL_MAX_OPEN_CONNS", 100),
		MaxIdleConns:    envInt("MYSQL_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: time.Hour,
	}); err != nil {
		log.Fatalf("mysql init failed: %v", err)
	}

	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8888"
	}

	h := server.Default(
		server.WithMaxRequestBodySize(500*1024*1024),
		server.WithHostPorts(host+":"+port),
	)

	// 添加 CORS 中间件
	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Refresh-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 配置静态文件服务 - 使用自定义处理函数
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "/app/uploads"
	}
	h.GET("/uploads/*filepath", func(c context.Context, ctx *app.RequestContext) {
		filepathStr := ctx.Param("filepath")
		// 清理路径，防止目录遍历攻击
		filepathStr = strings.TrimPrefix(filepathStr, "/")
		fullPath := filepath.Join(uploadsDir, filepathStr)

		// 安全检查：确保请求的文件在uploads目录内
		absPath, _ := filepath.Abs(fullPath)
		absUploadsDir, _ := filepath.Abs(uploadsDir)
		if !strings.HasPrefix(absPath, absUploadsDir) {
			ctx.String(consts.StatusForbidden, "Access denied")
			return
		}

		// 检查文件是否存在
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			ctx.String(consts.StatusNotFound, "File not found: "+filepathStr)
			return
		}

		// 提供文件服务
		ctx.File(fullPath)
	})

	register(h)
	h.Spin()
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("env %s is required", key)
	}
	return v
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
