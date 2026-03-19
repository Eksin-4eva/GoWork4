# BiliGO

仅完成了最低要求

## 技术栈

- **框架**: Cloudwego Hertz
- **ORM**: GORM + gorm/gen
- **数据库**: MySQL
- **认证**: JWT（access token + refresh token）
- **语言**: Golang

## 项目结构

```
.
├── main.go                  # 入口，服务器初始化
├── router.go / router_gen.go
├── biz/
│   ├── handler/api/         # HTTP 处理器
│   ├── service/             # 业务逻辑
│   ├── dal/
│   │   ├── model/           # GORM 模型
│   │   ├── query/           # gorm/gen 生成的查询
│   │   └── mysql/           # 数据库初始化
│   ├── mw/                  # 中间件（JWT、请求体大小限制）
│   ├── model/api/           # 请求/响应结构体
│   └── router/api/          # 路由注册
└── pkg/utils/               # JWT 工具
```

## API 接口

### 用户
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/user/register` | 否 | 注册 |
| POST | `/user/login` | 否 | 登录 |
| GET | `/user/info` | 是 | 获取用户信息 |
| PUT | `/user/avatar/upload` | 是 | 上传头像 |
| POST | `/user/refresh` | 否 | 刷新 Token |

### 视频
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/video/publish` | 是 | 发布视频（multipart） |
| GET | `/video/list` | 是 | 获取用户视频列表 |
| GET | `/video/popular` | 否 | 热门视频 |
| POST | `/video/search` | 否 | 关键词搜索视频 |

### 互动
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/like/action` | 是 | 点赞 / 取消点赞 |
| GET | `/like/list` | 是 | 获取点赞视频列表 |
| POST | `/comment/publish` | 是 | 发布评论 |
| GET | `/comment/list` | 否 | 获取视频评论列表 |
| DELETE | `/comment/delete` | 是 | 删除评论 |

### 社交
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/relation/action` | 是 | 关注 / 取消关注 |
| GET | `/following/list` | 是 | 获取关注列表 |
| GET | `/follower/list` | 是 | 获取粉丝列表 |
| GET | `/friends/list` | 是 | 获取互关好友列表 |

## 快速开始

### 环境要求

- Go 1.21+
- MySQL

### 配置

复制 `.env.example` 为 `.env` 并填写配置：

```env
MYSQL_DSN=user:password@tcp(localhost:3306)/biligo?charset=utf8mb4&parseTime=True
JWT_SECRET=your_secret
SERVER_HOST=0.0.0.0
SERVER_PORT=8888
UPLOADS_DIR=/app/uploads
```

### 本地运行

```bash
go run .
```

### Docker 部署

```bash
docker compose up -d
```

## 认证说明

需要认证的接口请在请求头中携带：

```
Authorization: Bearer <access_token>
```

刷新 Token 时，将 refresh token 放入 `X-Refresh-Token` 请求头，调用 `POST /user/refresh`。

## 文件上传

- 视频：最大 500MB，保存至 `uploads/videos/`
- 封面：最大 10MB，保存至 `uploads/covers/`
- 头像：保存至 `uploads/avatars/`

静态文件通过 `/uploads/*` 路径访问。
