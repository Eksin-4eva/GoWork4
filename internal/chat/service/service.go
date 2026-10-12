// Package service 实现 WebSocket 聊天服务。
package service

import (
	"net/http"

	"github.com/gorilla/websocket"

	"gobili/config"
	"gobili/pkg/base"
	"gobili/pkg/constants"
	"gobili/pkg/db"
	"gobili/pkg/errno"
	"gobili/pkg/refresh"
	"gobili/pkg/token"
	"gobili/pkg/utils"
)

// Service 是聊天服务的实现。
type Service struct {
	db        *db.Database
	sf        *utils.Snowflake
	hub       *hub
	upgrader  websocket.Upgrader
	exchanger refresh.Exchanger
}

// New 构造聊天服务。依赖取自进程级 ClientSet。
func New() *Service {
	cs := base.GetClientSet()
	return newService(cs.DB, cs.SF, config.Get().Chat.AllowedOrigins)
}

// newService 用给定依赖构造服务。
//
// 单独拆出来是为了让测试能绕过进程级 ClientSet——
// 聊天逻辑只需要数据库和雪花生成器，不该为了测试去连 Redis 和 MinIO。
func newService(database *db.Database, sf *utils.Snowflake, allowedOrigins []string) *Service {
	return &Service{
		db:        database,
		sf:        sf,
		hub:       newHub(),
		exchanger: refresh.Exchanger{DB: database, SF: sf},
		upgrader: websocket.Upgrader{
			CheckOrigin: originChecker(allowedOrigins),
		},
	}
}

// authenticate 校验握手/请求携带的令牌，并轮换一对新令牌。
//
// 令牌来源有两个：Authorization 请求头与 ?token= 查询参数。
// 后者是必需的——浏览器的 WebSocket API 不允许设置自定义请求头。
func (s *Service) authenticate(r *http.Request) (int64, map[string]string, error) {
	access := r.Header.Get(constants.AuthHeader)
	if access == "" {
		access = r.URL.Query().Get("token")
	}
	refresh := r.Header.Get(constants.RefreshTokenHeader)
	if refresh == "" {
		refresh = r.URL.Query().Get("refresh_token")
	}

	if access == "" && refresh == "" {
		return 0, nil, errno.AuthMissing
	}

	// 与 HTTP 网关共用同一条令牌规则（pkg/token），但换发实现在本进程内完成——
	// chat 自己连着数据库，不需要为此发一次 RPC。
	userID, newRefresh, err := token.ResolveUser(r.Context(), s.exchanger, access, refresh)
	if err != nil {
		return 0, nil, err
	}

	// 与 HTTP 网关保持同一套规则：换发新的访问令牌；
	// 刷新令牌只在这次是靠它通过的时候才换发（换发时已经落库）。
	newAccess, err := token.CreateToken(constants.TypeAccessToken, userID)
	if err != nil {
		return 0, nil, err
	}

	tokens := map[string]string{constants.AccessTokenHeader: newAccess}
	if newRefresh != "" {
		tokens[constants.RefreshTokenHeader] = newRefresh
	}
	return userID, tokens, nil
}

// originChecker 按白名单校验握手的 Origin 请求头。
//
// 旧版这里恒返回 true，等于允许任意站点用访客的身份建立连接。
// 白名单为空时只放行不带 Origin 的客户端（原生 App、curl、服务端调用）。
func originChecker(allowed []string) func(*http.Request) bool {
	whitelist := make(map[string]struct{}, len(allowed))
	for _, origin := range allowed {
		whitelist[origin] = struct{}{}
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		_, ok := whitelist[origin]
		return ok
	}
}
