iimi# ReBiliGOOO 实现步骤

1. 基于 go-zero 搭建了 `gateway` 服务，整理并落地了 `AIDocs` 中的视频站接口，统一生成了路由、handler、logic、types 等基础代码。
2. 扩展了网关配置，补充了 MySQL、JWT 双 token、上传目录、静态资源地址、Redis 等配置项，供后续业务层统一使用。
3. 设计并写入了数据库初始化脚本 `gateway/assets/sql/init.sql`，包含用户、刷新令牌、视频、评论、点赞、关注、图搜记录、聊天消息等核心表。
4. 新增 `internal/app` 业务层和 `repository` 数据访问层，把注册、登录、鉴权、用户信息、MFA、投稿、图搜、点赞、评论、关注、列表查询等逻辑集中到应用层处理。
5. 实现了 JWT 双 token 流程：登录时签发 Access/Refresh Token；鉴权时优先校验 Access Token，必要时使用 Refresh Token 换发新的 Access Token。
6. 增加了基于请求头的鉴权辅助逻辑，受保护接口会先注入认证用户，再进入对应 logic。
7. 修复了上传类接口的 handler 与 logic 签名不一致的问题，已经支持从 multipart/form-data 中读取文件并传递给业务层：
   - 用户头像上传
   - 视频投稿
   - 以图搜图
8. 修复了登录接口的 token 返回方式，当前会在响应头中写入配置里的 `Access-Token` 和 `Refresh-Token`。
9. 清理了 goctl 初始残留的无用示例 scaffold（如旧的 `GatewayHandler/GatewayLogic`），避免影响编译。
10. 给网关入口挂上了 `/static` 静态文件服务，当前会直接暴露 `gateway/uploads` 目录，已与 `Upload.BaseURL` 的返回地址对齐。
11. 补完了独立的 WebSocket `chat` 服务：
    - 新增 `chat/main.go` 启动入口
    - 新增 `chat/internal/config` 配置结构
    - 新增 `chat/internal/server` HTTP/WebSocket 路由装配
    - 新增 `chat/internal/service`，实现 JWT 鉴权、连接管理、消息落库、在线转发、发送回执
12. chat 服务复用了现有双 token 规则：握手时优先读取 `Access-Token`，必要时使用 `Refresh-Token` 校验并补发新的 Access Token。
13. 执行了依赖拉取与编译检查：
    - 使用 `GOPROXY=https://goproxy.cn,direct go mod tidy` 拉取依赖
    - 使用 `GOPROXY=https://goproxy.cn,direct go get github.com/gorilla/websocket` 增加 WebSocket 依赖
    - 使用 `GOPROXY=https://goproxy.cn,direct go build ./...` 完成整仓编译通过

## 当前代码重点

- 网关入口：`gateway/gateway.go`
- 配置结构：`gateway/internal/config/config.go`
- 业务核心：`gateway/internal/app/app.go`
- 数据访问：`gateway/internal/app/repository.go`
- 鉴权辅助：`gateway/internal/handler/auth_helper.go`
- 聊天服务入口：`chat/main.go`
- 聊天服务实现：`chat/internal/service/service.go`
- 数据库脚本：`gateway/assets/sql/init.sql`

## 说明

1. 代码目前已经能通过编译，但本机环境里之前没有 `mysql` 命令行工具，所以数据库脚本是否已真正执行到本地 MySQL，需要你在有 MySQL 客户端的环境中执行 `init.sql` 再验证。
2. chat 服务当前提供 `/ws` WebSocket 握手入口和 `/healthz` 健康检查；消息体采用 JSON，最小字段为 `to_user_id` 和 `content`，`message_type` 为空时默认按 `text` 处理。
3. 上传文件返回的 URL 现在会走 `http://127.0.0.1:8888/static/...`，前提是运行网关时工作目录位于 `gateway/` 下，使 `uploads` 相对路径正确映射到本地上传目录。
