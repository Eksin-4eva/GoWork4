package constants

import "time"

// 服务名。
const (
	ServiceAPI  = "gobili-api"
	ServiceChat = "gobili-chat"
)

// 分页默认值与保护上限，避免客户端传入超大 page_size 拖垮数据库。
const (
	DefaultPageNum  = 1
	DefaultPageSize = 10
	MaxPageSize     = 50
)

// 数据库连接池与超时配置。
const (
	DBSlowThreshold   = time.Second
	DBMaxIdleConns    = 10
	DBMaxOpenConns    = 100
	DBConnMaxLifetime = time.Hour
)

// 业务约束。
const (
	// MaxUploadMB 是单个上传文件的大小上限（MB）。
	MaxUploadMB = 100
	// FeedLimit 是 feed 流单次返回的条数。
	FeedLimit = 30
)
