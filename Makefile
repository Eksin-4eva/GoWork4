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

.PHONY: run-api
run-api:
	go run ./cmd/api -f "$(CONFIG_PATH)/config.yaml"

## --------------------------------------
## 测试
## --------------------------------------

.PHONY: test
test:
	go test -v -gcflags="all=-l -N" -race -coverprofile=coverage.txt -parallel=16 -p=16 \
		-covermode=atomic -coverpkg=./... \
		`go list ./... | grep -E -v "idl|config|docs|docker"`

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
