# BiliGO

Golang第四轮作业

## 技术栈

- **框架**: Cloudwego Hertz
- **ORM**: GORM + gorm/gen
- **数据库**: MySQL
- **对象存储**: MinIO
- **认证**: JWT（access token + refresh token）+ MFA
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
│   │   ├── mysql/           # 数据库初始化
│   │   └── minio/           # MinIO 对象存储
│   ├── mw/                  # 中间件（JWT、请求体大小限制）
│   ├── model/api/           # 请求/响应结构体
│   └── router/api/          # 路由注册
├── idl/                     # Thrift IDL 定义
│   ├── common.thrift        # 公共类型
│   ├── user/                # 用户模块
│   ├── video/               # 视频模块
│   ├── interact/            # 互动模块
│   ├── relation/            # 社交模块
│   └── auth/                # 认证模块
└── pkg/utils/               # JWT 工具
```

## 完成情况

### 最低要求（17个接口）✅

| 模块 | 接口 | 状态 |
|------|------|------|
| 用户 | 注册、登录、用户信息、上传头像 | ✅ 4/4 |
| 视频 | 投稿、发布列表、搜索视频、热门排行榜 | ✅ 4/4 |
| 互动 | 点赞操作、点赞列表、评论、评论列表、删除评论 | ✅ 5/5 |
| 社交 | 关注操作、关注列表、粉丝列表、好友列表 | ✅ 4/4 |

### 额外功能

| 功能 | 状态 | 说明 |
|------|------|------|
| 双Token认证 | ✅ | access_token + refresh_token |
| MFA多因素认证 | ✅ | TOTP绑定、二维码获取 |
| 评论点赞 | ✅ | 支持对评论点赞 |
| 嵌套评论 | ✅ | 支持对评论进行评论 |
| MinIO存储 | ✅ | 视频文件对象存储 |
| Docker部署 | ✅ | Dockerfile + docker-compose |

## API 接口

### 用户模块
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/user/register` | 否 | 注册 |
| POST | `/user/login` | 否 | 登录 |
| GET | `/user/info` | 是 | 获取用户信息 |
| PUT | `/user/avatar/upload` | 是 | 上传头像 |
| POST | `/user/refresh` | 否 | 刷新 Token |
| GET | `/auth/mfa/qrcode` | 是 | 获取MFA二维码 |
| POST | `/auth/mfa/bind` | 是 | 绑定MFA |

### 视频模块
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/video/publish` | 是 | 发布视频（multipart） |
| GET | `/video/list` | 否 | 获取用户视频列表 |
| GET | `/video/popular` | 否 | 热门视频排行榜 |
| POST | `/video/search` | 否 | 关键词搜索视频 |

### 互动模块
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/like/action` | 是 | 点赞/取消点赞（支持视频和评论） |
| GET | `/like/list` | 否 | 获取点赞视频列表 |
| POST | `/comment/publish` | 是 | 发布评论（支持视频和评论） |
| GET | `/comment/list` | 否 | 获取评论列表（支持子评论） |
| DELETE | `/comment/delete` | 是 | 删除评论 |

### 社交模块
| 方法 | 路径 | 需要认证 | 说明 |
|------|------|----------|------|
| POST | `/relation/action` | 是 | 关注/取消关注 |
| GET | `/following/list` | 否 | 获取关注列表 |
| GET | `/follower/list` | 否 | 获取粉丝列表 |
| GET | `/friends/list` | 否 | 获取互关好友列表 |

## 快速开始

### 环境要求

- Go 1.21+
- MySQL
- MinIO (可选，用于对象存储)

### 配置

复制 `.env.example` 为 `.env` 并填写配置：

```env
MYSQL_DSN=root:password@tcp(localhost:3306)/biligo?charset=utf8mb4&parseTime=True&loc=Local
MYSQL_MAX_OPEN_CONNS=100
MYSQL_MAX_IDLE_CONNS=10
JWT_SECRET=your_jwt_secret

# MinIO 配置
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET=biligo
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
Access-Token: <access_token>
Refresh-Token: <refresh_token>
```

刷新 Token 时，调用 `POST /user/refresh`。

## 文件上传

- 视频：最大 500MB，存储至 MinIO
- 封面：最大 10MB，存储至 MinIO
- 头像：存储至 MinIO

## 数据库设计

### 用户表 (users)
- id, username, password, avatar_url, mfa_secret
- follower_count, following_count

### 视频表 (videos)
- id, user_id, title, description
- video_url, cover_url
- view_count, like_count, comment_count

### 评论表 (comments)
- id, video_id, user_id, content, parent_id
- like_count, child_count

### 点赞表 (favorites)
- id, user_id, video_id

### 评论点赞表 (comment_likes)
- id, user_id, comment_id

### 关系表 (relations)
- id, user_id, to_user_id, status
