# Gobili - 视频网站 API

基于 go-zero 框架开发的视频网站后端 API 服务，完成 FZU West2 Golang Lab4 作业要求。

## 项目结构

```
GoWork4/
├── Dockerfile                  # 多阶段构建（同时编译 gateway + chat）
├── docker-compose.yml          # Docker 编排（mysql+redis+app）
├── gateway/                    # API 网关服务（主服务）
│   ├── gateway.api             # go-zero API 定义文件（goctl 脚手架）
│   ├── gateway.go              # 入口 main
│   ├── etc/
│   │   ├── gateway-api.yaml    # 本地配置
│   │   └── gateway-api-docker.yaml  # Docker 配置
│   ├── assets/sql/
│   │   └── init.sql            # 数据库初始化 SQL
│   ├── internal/
│   │   ├── config/config.go    # 配置结构体
│   │   ├── handler/            # 路由 handler（goctl 生成 + 手动补充）
│   │   │   ├── routes.go       # 路由注册（goctl 生成）
│   │   │   ├── auth_helper.go  # 鉴权辅助
│   │   │   └── *_handler.go    # 各接口 handler
│   │   ├── logic/              # 业务逻辑层
│   │   ├── middleware/         # 中间件（Auth）
│   │   ├── svc/                # 服务上下文
│   │   ├── types/types.go      # 请求/响应类型（goctl 生成）
│   │   └── app/
│   │       ├── app.go          # 核心业务逻辑 + 模型定义
│   │       └── repository.go   # 数据库访问层
│   └── uploads/                # 文件上传目录
├── chat/                       # WebSocket 聊天服务（独立进程）
│   ├── main.go
│   ├── etc/
│   │   ├── chat.yaml           # 本地配置
│   │   └── chat-docker.yaml    # Docker 配置
│   └── internal/
│       ├── config/config.go
│       ├── server/server.go
│       └── service/service.go
├── pkg/
│   └── migration/migration.go  # 数据库自动迁移
├── go.mod
└── go.sum
```

## 技术栈

- 框架：go-zero（使用 goctl 生成脚手架代码）
- 数据库：MySQL
- 缓存：Redis（go-redis/v9）
- 认证：JWT 双 Token（Access Token + Refresh Token）
- WebSocket：gorilla/websocket
- MFA：TOTP 二步验证


## 已完成的必做接口（17/17）

### 用户模块（4/4）
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 注册 | POST | /user/register | ✅ |
| 登录 | POST | /user/login | ✅ |
| 用户信息 | GET | /user/info | ✅ |
| 上传头像 | PUT | /user/avatar/upload | ✅ |

### 视频模块（4/4）
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 投稿 | POST | /video/publish | ✅ |
| 发布列表 | GET | /video/list | ✅ |
| 搜索视频 | POST | /video/search | ✅ |
| 热门排行榜 | GET | /video/popular | ✅ |

### 互动模块（5/5）
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 点赞操作 | POST | /like/action | ✅ |
| 点赞列表 | GET | /like/list | ✅ |
| 评论 | POST | /comment/publish | ✅ |
| 评论列表 | GET | /comment/list | ✅ |
| 删除评论 | DELETE | /comment/delete | ✅ |

### 社交模块（4/4）
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 关注操作 | POST | /relation/action | ✅ |
| 关注列表 | GET | /following/list | ✅ |
| 粉丝列表 | GET | /follower/list | ✅ |
| 好友列表 | GET | /friends/list | ✅ |

## 必做要求完成情况

| 要求 | 状态 | 说明 |
|------|------|------|
| 分页管理 | ✅ | 所有带 page_num/page_size 的接口均已实现分页 |
| 视频搜索多条件 | ✅ | 支持 keywords、from_date、to_date、username 组合查询 |
| 删除评论权限校验 | ✅ | 仅允许评论作者删除自己的评论 |
| 双 Token 认证 | ✅ | Access Token（2h）+ Refresh Token（7d），Access 过期自动用 Refresh 刷新 |
| 使用现代 HTTP 框架 | ✅ | 使用 go-zero，通过 goctl 生成脚手架代码 |
| 项目结构图 | ✅ | 见上方目录树 |
| Docker 部署 | ✅ | Dockerfile + docker-compose 一键部署 |
| 请求/返回结构遵循文档 | ✅ | 统一 `{base: {code, msg}, data: ...}` 格式 |

## Bonus 完成情况

| Bonus 项 | 状态 | 说明                                              |
|-----------|--|-------------------------------------------------|
| 实现全部接口 | ✅ | 除了以图搜图，不是很明白这个接口用来干什么                           |
| Redis 缓存点赞 | ❌ | 点赞操作直接走数据库                                      |
| Redis 热门排行榜 | ✅ | 使用 Redis Sorted Set 缓存热门视频排行，5 分钟过期，访问视频时实时更新分数 |
| 分片上传 | ❌ | 未实现                                             |
| WebSocket 聊天 | ✅ | 独立 chat 服务，支持 1v1 实时聊天、消息持久化、在线状态感知             |
| ElasticSearch 日志 | ❌ | 未引入                                             |

## 数据库设计

共 7 张表：users、refresh_tokens、videos、comments、likes、relations、chat_messages

- 使用雪花 ID 风格的字符串主键
- 软删除（deleted_at）
- 合理的索引设计（外键索引、全文索引、唯一约束）
- 启动时自动迁移建表

## 快速启动

### 本地开发

```bash
# 前置依赖：MySQL、Redis

# 启动 API 网关
cd gateway && go run gateway.go

# 启动聊天服务（可选）
cd chat && go run main.go
```

### Docker 部署

```bash
docker-compose up --build
```

自动完成 MySQL/Redis 启动、数据库建表、服务编译部署。

- gateway: http://localhost:8888
- chat: http://localhost:8889
