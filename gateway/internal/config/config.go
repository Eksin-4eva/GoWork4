package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth   AuthConfig
	MySQL  MySQLConfig
	JWT    JWTConfig
	Upload UploadConfig
	Redis  RedisConfig
}

type AuthConfig struct {
	AccessHeader  string
	RefreshHeader string
}

type MySQLConfig struct {
	DSN string
}

type JWTConfig struct {
	AccessSecret         string
	RefreshSecret        string
	AccessExpireSeconds  int64
	RefreshExpireSeconds int64
}

type UploadConfig struct {
	BaseURL    string
	AvatarDir  string
	VideoDir   string
	ImageDir   string
	MaxMB      int64
}

type RedisConfig struct {
	Addr string
}
