# gobili

仿抖音视频站后端，基于 CloudWeGo **Hertz** 构建。

本仓库由 GoWork4（go-zero 版）重构而来，工程规范对齐 `fzuhelper-server`。
迁移方案见仓库上级目录的 `HERTZ_MIGRATION_PLAN.md`。

## 目录结构

```
.
├── api              # HTTP 层：handler / model / pack / mw / router
├── cmd              # 各服务入口
│   └── api
├── idl              # thrift 接口定义
├── internal         # 业务实现（按领域拆分）
└── pkg              # 基础设施：base / cache / db / constants / errno / logger / utils
```

## 常用命令

```bash
make help            # 查看全部命令
make hertz-gen-api   # 由 idl/api.thrift 生成脚手架
make build-api       # 构建 api 服务
make verify          # fmt + import + vet + lint
make test            # 单元测试
```

## 当前进度

WORK IN PROGRESS

## 本地开发

```bash
# 1. 启动依赖（MySQL / Redis / MinIO）
docker compose -f docker/docker-compose.yml up -d

# 2. 准备本地配置
cp config/config.example.yaml config/config.yaml

# 3. 启动服务
make run-api
```

MinIO 控制台：http://127.0.0.1:9001 （gobili / gobili123）
