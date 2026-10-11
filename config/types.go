// Package config 负责加载与访问服务配置。
package config

import (
	"fmt"
	"net/url"
	"time"
)

// Config 是配置文件的根结构。
type Config struct {
	Server    Server
	MySQL     MySQL
	Redis     Redis
	MinIO     MinIO
	JWT       JWT
	Snowflake Snowflake
	Chat      Chat
	Kafka     Kafka
	Etcd      Etcd
	Services  Services
}

// Server 是 HTTP 服务自身的基础配置。
type Server struct {
	Name     string
	Host     string
	Port     int
	LogLevel string
}

// MySQL 是数据库连接配置。
type MySQL struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Charset  string
	// Loc 决定驱动如何解析 DATETIME 列，取值是 Go 的时区名（如 Asia/Shanghai）。
	Loc string
	// TimeZone 是数据库会话时区，取值是 SQL 里能识别的偏移（如 +08:00）。
	// 它必须与 Loc 表达同一个时区，否则 created_at（应用写入）
	// 与 updated_at（由 CURRENT_TIMESTAMP 生成）会相差几个小时。
	TimeZone string
}

// Redis 是缓存连接配置。
type Redis struct {
	Addr     string
	Password string
	DB       int
}

// MinIO 是对象存储配置。
type MinIO struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	PublicBaseURL string
}

// JWT 是双令牌签名与有效期配置。
type JWT struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// Snowflake 是雪花 ID 生成器的节点配置。
type Snowflake struct {
	DatacenterID int64
	WorkerID     int64
}

// Chat 是 WebSocket 聊天服务自身的配置。
type Chat struct {
	Host string
	Port int
	// AllowedOrigins 是允许的握手来源白名单。
	//
	// 留空表示只接受不带 Origin 的客户端（原生 App、curl、服务端调用），
	// 浏览器页面会被拒绝——这是安全的默认值：Origin 校验形同虚设的后果是
	// 任何网站都能拿访客的身份连上来。
	AllowedOrigins []string
}

// Kafka 是消息队列配置。
type Kafka struct {
	Brokers []string
	GroupID string
	// FfmpegPath 是 ffmpeg 可执行文件路径，留空则用 PATH 里的 ffmpeg。
	FfmpegPath string
}

// Etcd 是服务注册与发现的配置。
type Etcd struct {
	Endpoints []string
}

// Services 是各微服务的注册名与监听地址。
//
// 网关只用 Name 去 etcd 里找服务；Host/Port 是服务端自己监听用的，
// 也就是说**网关不关心服务部署在哪台机器上**——地址由 etcd 动态给出。
type Services struct {
	User    ServiceEndpoint
	Video   ServiceEndpoint
	Comment ServiceEndpoint
}

// ServiceEndpoint 是一个微服务的注册名与监听地址。
type ServiceEndpoint struct {
	// Name 是注册到 etcd 的服务名，客户端靠它发现服务
	Name string
	Host string
	Port int
}

// DSN 拼接 MySQL 连接串。
func (m MySQL) DSN() string {
	params := url.Values{}
	params.Set("charset", m.Charset)
	params.Set("parseTime", "true")
	params.Set("loc", m.Loc)
	// time_zone 会被驱动当成会话变量下发（SET time_zone='+08:00'），
	// 值需要写成 SQL 字符串字面量，url 编码后是 %27+08%3A00%27
	params.Set("time_zone", "'"+m.TimeZone+"'")

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		m.Username, m.Password, m.Host, m.Port, m.Database, params.Encode())
}
