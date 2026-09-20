// Package config 负责加载与访问服务配置。
package config

import (
	"fmt"
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

// DSN 拼接 MySQL 连接串。
func (m MySQL) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		m.Username, m.Password, m.Host, m.Port, m.Database, m.Charset,
	)
}
