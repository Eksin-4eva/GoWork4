# 辅助工具安装
#   go install github.com/cloudwego/hertz/cmd/hz@latest
#   go install golang.org/x/tools/cmd/goimports@latest
#   go install mvdan.cc/gofumpt@latest
#   golangci-lint: https://golangci-lint.run/welcome/install/

.DEFAULT_GOAL := help

# 项目 MODULE 名
MODULE = gobili
# 目录相关
DIR = $(shell pwd)
IDL_PATH = $(DIR)/idl
CONFIG_PATH = $(DIR)/config
OUTPUT_PATH = $(DIR)/output

# 工具链：默认取 GOPATH/bin，避免依赖 shell PATH
GOBIN_PATH = $(shell go env GOPATH)/bin
HZ = $(GOBIN_PATH)/hz
GOFUMPT = $(GOBIN_PATH)/gofumpt
GOIMPORTS = $(GOBIN_PATH)/goimports
GOLANGCI_LINT = $(GOBIN_PATH)/golangci-lint

PREFIX = "[Makefile]"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  hertz-gen-api  : Generate Hertz scaffold from idl/api.thrift"
	@echo "  build-api      : Build the api service into output/bin"
	@echo "  build-chat     : Build the chat service into output/bin"
	@echo "  build-worker   : Build the worker service into output/bin"
	@echo "  build-rpc      : Build the three microservices (user/video/comment)"
	@echo "  run-api        : Run the api service locally"
	@echo "  test           : Run unit tests with coverage"
	@echo "  fmt            : Format code with gofumpt"
	@echo "  import         : Optimize imports with goimports"
	@echo "  vet            : Run go vet"
	@echo "  lint           : Run golangci-lint"
	@echo "  verify         : fmt + import + vet + lint"
	@echo "  clean          : Remove output/ and coverage.txt"

## --------------------------------------
## 代码生成
## --------------------------------------

# 生成基于 Hertz 的脚手架
# 注意：api.thrift 中涉及 multipart 的接口，生成后需要手工把 model 里的字段
# 改成 *multipart.FileHeader，重跑本命令会覆盖这些改动。
.PHONY: hertz-gen-api
hertz-gen-api:
	PATH="$(GOBIN_PATH):$$PATH" $(HZ) update -idl "$(IDL_PATH)/api.thrift" -t template=slim
	go mod tidy

## --------------------------------------
## 构建与运行
## --------------------------------------

.PHONY: build-api
build-api:
	@mkdir -p "$(OUTPUT_PATH)/bin"
	CGO_ENABLED=0 go build -o "$(OUTPUT_PATH)/bin/api" ./cmd/api

.PHONY: build-chat
build-chat:
	@mkdir -p "$(OUTPUT_PATH)/bin"
	CGO_ENABLED=0 go build -o "$(OUTPUT_PATH)/bin/chat" ./cmd/chat

.PHONY: build-worker
build-worker:
	@mkdir -p "$(OUTPUT_PATH)/bin"
	CGO_ENABLED=0 go build -o "$(OUTPUT_PATH)/bin/worker" ./cmd/worker

# 三个 RPC 微服务
.PHONY: build-rpc
build-rpc: build-user build-video build-comment

.PHONY: build-user
build-user:
	@mkdir -p "$(OUTPUT_PATH)/bin"
	CGO_ENABLED=0 go build -o "$(OUTPUT_PATH)/bin/user" ./cmd/user

.PHONY: build-video
build-video:
	@mkdir -p "$(OUTPUT_PATH)/bin"
	CGO_ENABLED=0 go build -o "$(OUTPUT_PATH)/bin/video" ./cmd/video

.PHONY: build-comment
build-comment:
	@mkdir -p "$(OUTPUT_PATH)/bin"
	CGO_ENABLED=0 go build -o "$(OUTPUT_PATH)/bin/comment" ./cmd/comment

.PHONY: build-all
build-all: build-api build-chat build-worker build-rpc

.PHONY: run-api
run-api:
	go run ./cmd/api -f "$(CONFIG_PATH)/config.yaml"

.PHONY: run-chat
run-chat:
	go run ./cmd/chat -f "$(CONFIG_PATH)/config.yaml"

.PHONY: run-worker
run-worker:
	go run ./cmd/worker -f "$(CONFIG_PATH)/config.yaml"

# 三个 RPC 微服务需要分别启动（各自一个终端）
.PHONY: run-user
run-user:
	go run ./cmd/user -f "$(CONFIG_PATH)/config.yaml"

.PHONY: run-video
run-video:
	go run ./cmd/video -f "$(CONFIG_PATH)/config.yaml"

.PHONY: run-comment
run-comment:
	go run ./cmd/comment -f "$(CONFIG_PATH)/config.yaml"

## --------------------------------------
## IDL 代码生成
## --------------------------------------

# 由 thrift IDL 生成 Kitex 的客户端/服务端代码。
#
# 只生成到 kitex_gen，不生成脚手架（server 的 main 与 handler 都是手写的，
# 生成器给的骨架反而要删）。新增方法后重跑本命令即可。
.PHONY: kitex-gen
kitex-gen:
	PATH="$(GOBIN_PATH):$$PATH" kitex -module $(MODULE) -thrift template=slim "$(IDL_PATH)/user.thrift"
	PATH="$(GOBIN_PATH):$$PATH" kitex -module $(MODULE) -thrift template=slim "$(IDL_PATH)/video.thrift"
	PATH="$(GOBIN_PATH):$$PATH" kitex -module $(MODULE) -thrift template=slim "$(IDL_PATH)/comment.thrift"
	go mod tidy

## --------------------------------------
## 测试
## --------------------------------------

.PHONY: test
test:
	go test -v -gcflags="all=-l -N" -race -coverprofile=coverage.txt -parallel=16 -p=16 \
		-covermode=atomic -coverpkg=./... \
		`go list ./... | grep -E -v "idl|config|docs|docker"`

# 用 config/sql/init.sql 重建独立的测试库，并给应用账号授权。
# 集成测试会 TRUNCATE 表，所以绝不能指向开发库。
.PHONY: test-db
test-db:
	docker exec gobili-mysql mysql -uroot -proot123456 -e "DROP DATABASE IF EXISTS gobili_test; CREATE DATABASE gobili_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci; GRANT ALL PRIVILEGES ON gobili_test.* TO 'gobili'@'%'; FLUSH PRIVILEGES;"
	docker exec -i gobili-mysql mysql -uroot -proot123456 gobili_test < "$(DIR)/config/sql/init.sql"

# 需要真实数据库的集成测试（DAL + 聊天服务）。未起依赖时会全部跳过。
.PHONY: test-integration
test-integration:
	GOBILI_TEST_DSN='gobili:gobili@tcp(127.0.0.1:3306)/gobili_test?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai' \
		go test -count=1 -race -v ./pkg/db/... ./internal/...

# 端到端冒烟测试。需要依赖容器与 api 服务都已经在跑。
.PHONY: smoke
smoke:
	bash "$(DIR)/hack/smoke.sh"

## --------------------------------------
## 代码规范
## --------------------------------------

.PHONY: fmt
fmt:
	$(GOFUMPT) -l -w .

.PHONY: import
import:
	$(GOIMPORTS) -l -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: verify
verify: fmt import vet lint

## --------------------------------------
## 清理
## --------------------------------------

.PHONY: clean
clean:
	rm -rf "$(OUTPUT_PATH)" coverage.txt
