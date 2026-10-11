# gobili

仿抖音视频站后端。基于 CloudWeGo **Hertz（网关）+ Kitex（微服务）+ etcd（服务发现）** 构建，
工程规范对齐 `west2-online/fzuhelper-server`。

本仓库由 GoWork4（go-zero 版）重构而来。重构方案与每一阶段的实施记录见仓库上级目录的
`HERTZ_MIGRATION_PLAN.md`。

## 技术栈

| 组件 | 选型 | 用途 |
|---|---|---|
| HTTP 网关 | Hertz + `hz` + thrift IDL | 对外 HTTP，只做协议适配 |
| RPC 框架 | **Kitex** + thrift IDL | 三个微服务之间的调用 |
| 服务发现 | **etcd** | 微服务注册与动态发现 |
| 数据库 | MySQL 8 + GORM | 业务数据，DDL 手写不用 AutoMigrate |
| 缓存 | Redis 7 | 热门榜（ZSet）、事件去重 |
| 对象存储 | MinIO | 头像 / 视频 / 封面，库里存 object key |
| 消息队列 | Kafka（KRaft） | 封面异步生成、播放计数聚合 |
| 鉴权 | 双 Token（JWT HS256，带 jti） | Access 2h + Refresh 7d，每次请求轮换 |
| 视频抽帧 | ffmpeg | worker 生成封面 |

## 进程与数据流

```
                            ┌──────────────┐
   HTTP :8888  ────────────►│     api      │──────► MinIO（文件在这里上传）
                            └───────┬──────┘
                                    │ Kitex RPC
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
            ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
            │  user  :8881│ │ video :8882 │ │comment:8883 │
            └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
                   │               │               │
                   ▼               ▼               ▼
                MySQL      MySQL+Redis+Kafka     MySQL
                                   │
                                   │ 异步事件
                                   ▼
                            ┌─────────────┐
                            │   Kafka     │
                            └──────┬──────┘
                                   │ 消费
                                   ▼
                            ┌─────────────┐
                            │   worker    │──► MySQL / Redis / MinIO
                            └─────────────┘

   WS :8889  ──────────────►│    chat     │──────► MySQL
                            └─────────────┘

                              ┌─────────────┐
                              │    etcd     │  ← 三个微服务注册到这里
                              └─────────────┘     网关按服务名动态发现
```

**六个进程，各自的依赖被压到最少：**

| 进程 | 端口 | 需要什么 | 为什么 |
|---|---|---|---|
| api | 8888 | MinIO + etcd | 文件在这里上传，RPC 只传 key |
| user | 8881 | MySQL | 只做用户与令牌 |
| video | 8882 | MySQL + Redis + Kafka | 热榜、封面事件 |
| comment | 8883 | MySQL | 只做评论 |
| chat | 8889 | MySQL | WebSocket 长连接 |
| worker | — | 全部 | 抽帧 + 计数聚合 |

**「依赖最少」不是巧合**：用户服务的 `NewClientSet` 只传了 `WithMySQL()`，
所以**Redis 挂了它也照常启动**。如果一开始就一次性建齐所有依赖，
微服务化就失去了意义——它们不过是同一个大进程拆成几个壳。

## 目录结构

```
.
├── api/              HTTP 层：handler / model / pack / mw / router / **rpc（Kitex 客户端）**
├── cmd/              六个可执行入口：api / user / video / comment / chat / worker
├── config/           配置结构与示例；sql/init.sql 是全项目唯一的建表来源
├── docker/           Dockerfile 与 compose（mysql + redis + minio + kafka + api/chat/worker）
├── hack/smoke.sh     端到端冒烟测试，40 条断言覆盖 15 个接口的正常与异常路径
├── idl/              thrift 接口定义（api 是 HTTP 契约，其余是 RPC 契约）
├── kitex_gen/        Kitex 生成的 RPC 代码
├── internal/         业务实现：user / video / comment（微服务）/ chat / worker / event
└── pkg/              基础设施：base（ClientSet）/ cache / db / errno / kafka / logger / storage / token / utils
```

## 接口一览

统一响应信封：`{"code": "0", "message": "Success", "data": {...}}`

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/user/register` | — | 注册 |
| POST | `/user/login` | — | 登录，双 Token 走响应头 |
| GET | `/user/info` | ✅ | 用户信息 |
| PUT | `/user/avatar/upload` | ✅ | 上传头像 |
| GET | `/auth/mfa/qrcode` | ✅ | 获取 MFA 绑定二维码 |
| POST | `/auth/mfa/bind` | ✅ | 绑定 MFA |
| GET | `/video/feed` | — | 最新视频流 |
| POST | `/video/publish` | ✅ | 投稿（封面可选，不传则异步抽帧） |
| GET | `/video/list` | — | 某用户的投稿列表 |
| GET | `/video/popular` | — | 热门榜 |
| POST | `/video/search` | — | 多条件搜索 |
| POST | `/video/visit` | ✅ | 记录播放 |
| POST | `/comment/publish` | ✅ | 发评论 / 回复 |
| GET | `/comment/list` | — | 评论列表 |
| DELETE | `/comment/delete` | ✅ | 删除自己的评论 |

聊天服务（8889）：`GET /ws`（WebSocket）、`GET /history`（会话历史）、`GET /healthz`。

## 快速开始

### 方式一：容器一键起（推荐）

```bash
docker compose -f docker/docker-compose.yml up -d --build
bash hack/smoke.sh      # 40 条断言，全绿说明栈是通的
```

会起 10 个容器：mysql、redis、minio、kafka、**etcd**、
api（8888）、**user（8881）、video（8882）、comment（8883）**、chat（8889）、worker。
首次启动时 MySQL 会自动执行 `config/sql/init.sql` 建表。

### 方式二：宿主机开发

```bash
docker compose -f docker/docker-compose.yml up -d mysql redis minio kafka
cp config/config.example.yaml config/config.yaml

# 三个微服务（各开一个终端）
make run-user       # 8881
make run-video      # 8882
make run-comment    # 8883

# 网关与外围（各开一个终端）
make run-api        # 8888
make run-chat       # 8889
make run-worker     # Kafka 消费者
```

MinIO 控制台 http://127.0.0.1:9001 （gobili / gobili123）

## 常用命令

```bash
make help              # 全部命令
make verify            # fmt + import + vet + lint
make test              # 单元测试（不需要外部依赖）
make test-db           # 用 init.sql 重建独立测试库 gobili_test
make test-integration  # 需要真实 MySQL/Redis/MinIO 的集成测试
make smoke             # 端到端冒烟（需要栈已启动）
make hertz-gen-api     # 由 idl/api.thrift 重新生成 handler/model/router
```

`make test` 不需要任何外部依赖：连库的测试在没设 `GOBILI_TEST_DSN` 时会自动跳过。

## 几个设计决定

**DDL 手写，不用 AutoMigrate。** 结构定义与程序代码是两种东西；`config/sql/init.sql`
是全项目唯一的建表来源，评审时一眼能看到全貌。AutoMigrate 只能加列，删列与改类型都不安全。

**对外 URL 由 object key 拼出。** 库里存 `video/123.mp4` 而不是完整 URL，
换存储或换域名不用刷数据。封面与头像的桶设成匿名可读——视频站这些本来就是公开资源。

**错误码只分大类。** 参数 / 鉴权 / 业务 / 内部四类 + 一个成功码，细分场景复用同一个码、
靠文案区分，避免错误码随业务增长而膨胀。对外绝不回显原始错误。

**播放计数走异步聚合。** 同步写库时每次播放都要更新同一行，是全站最热的写路径。
现在只发事件，worker 攒批提交；消息队列不可用时自动降级为同步写，宁可退化也不静默丢计数。
