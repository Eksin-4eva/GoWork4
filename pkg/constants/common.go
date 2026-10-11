package constants

import "time"

// 服务名。
const (
	ServiceAPI    = "gobili-api"
	ServiceChat   = "gobili-chat"
	ServiceWorker = "gobili-worker"
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
	// PopularRankSize 是热门榜维护的条目数上限，
	// 比单页上限大得多，避免翻几页就翻空了。
	PopularRankSize = 200

	// 下列上限与 config/sql/init.sql 的列宽对应，超出会在写库时直接报错，
	// 所以在业务层先拦住并给出可读提示。
	MaxVideoTitleRunes = 128 // videos.title varchar(128)
	MaxCommentRunes    = 500 // comments.content text，业务上限

	// 聊天相关上限。
	MaxMessageRunes     = 2000 // 单条消息长度
	DefaultMessageType  = "text"
	ChatHistoryPageSize = 20
)
