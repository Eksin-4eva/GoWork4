package config

type Config struct {
	Name  string
	Host  string
	Port  int
	Auth  AuthConfig
	MySQL MySQLConfig
	JWT   JWTConfig
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
