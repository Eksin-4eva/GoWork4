package constants

import "time"

// Token 类型、有效期与请求/响应头名称。
const (
	TypeAccessToken  = 0
	TypeRefreshToken = 1

	AccessTokenTTL  = 2 * time.Hour
	RefreshTokenTTL = 7 * 24 * time.Hour

	// AuthHeader 是客户端携带令牌的请求头。
	AuthHeader = "Authorization"
	// AccessTokenHeader / RefreshTokenHeader 是服务端下发令牌的响应头。
	AccessTokenHeader  = "Access-Token"
	RefreshTokenHeader = "Refresh-Token"

	// UserIDContextKey 是鉴权中间件写入、业务层读取当前用户 ID 的上下文键。
	UserIDContextKey = "user_id"
)
