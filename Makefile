SHELL := /bin/sh

GO ?= go

GOPATH := $(shell $(GO) env GOPATH)
GOBIN := $(shell $(GO) env GOBIN)
ifeq ($(GOBIN),)
BIN_DIR := $(GOPATH)/bin
else
BIN_DIR := $(GOBIN)
endif

GOIMPORTS := $(BIN_DIR)/goimports
GOLANGCI := $(BIN_DIR)/golangci-lint
MIGRATE := $(BIN_DIR)/migrate

# 本地密钥/配置从 deploy/.env 读取（该文件不入库，见 .gitignore）
-include deploy/.env

MYSQL_HOST ?= 127.0.0.1
MYSQL_PORT ?= 3307
MYSQL_DB ?= seckill
MYSQL_USER ?= root
MYSQL_DSN ?= mysql://$(MYSQL_USER):$(MYSQL_ROOT_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DB)

COMPOSE := docker compose -f deploy/docker-compose.yml --env-file deploy/.env

TOOL_GOLANGCI := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
TOOL_GOIMPORTS := golang.org/x/tools/cmd/goimports@v0.49.0
TOOL_MIGRATE := github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

.PHONY: help tools up down logs ps migrate migrate-down migrate-force migrate-create fmt fmt-check vet golangci-lint lint test build tidy seed warmup run loadtest

help: ## 列出可用目标
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

tools: ## 安装钉死版本的工具链到 $(BIN_DIR)
	$(GO) install $(TOOL_GOLANGCI)
	$(GO) install $(TOOL_GOIMPORTS)
	$(GO) install -tags 'mysql' $(TOOL_MIGRATE)
	@echo "工具已安装到 $(BIN_DIR)"

up: ## 启动 mysql/redis（compose）
	$(COMPOSE) up -d
	$(COMPOSE) ps

down: ## 停止并移除容器（保留数据卷）
	$(COMPOSE) down

logs: ## 跟随查看 compose 日志
	$(COMPOSE) logs -f

ps: ## 查看 compose 容器状态
	$(COMPOSE) ps

migrate: ## 执行 migrations 到最新版本
	@test -n "$(MYSQL_ROOT_PASSWORD)" || { echo "MYSQL_ROOT_PASSWORD 未设置：请在 deploy/.env 中配置" >&2; exit 1; }
	@$(MIGRATE) -path migrations -database "$(MYSQL_DSN)" up

migrate-down: ## 回滚一个迁移版本
	@test -n "$(MYSQL_ROOT_PASSWORD)" || { echo "MYSQL_ROOT_PASSWORD 未设置：请在 deploy/.env 中配置" >&2; exit 1; }
	@$(MIGRATE) -path migrations -database "$(MYSQL_DSN)" down 1

migrate-force: ## 强制设置迁移版本并清除 dirty，用法：make migrate-force V=0
	@test -n "$(V)" || { echo "用法: make migrate-force V=0" >&2; exit 1; }
	@$(MIGRATE) -path migrations -database "$(MYSQL_DSN)" force $(V)

migrate-create: ## 新建迁移，用法：make migrate-create NAME=create_orders
	@test -n "$(NAME)" || { echo "用法: make migrate-create NAME=create_orders" >&2; exit 1; }
	$(MIGRATE) create -ext sql -dir migrations -seq $(NAME)

fmt: ## 格式化（gofmt + goimports）
	gofmt -w .
	$(GOIMPORTS) -w .

fmt-check: ## 检查格式，不修改文件
	@out=$$(gofmt -l .); test -z "$$out" || { echo "$$out"; echo "存在未格式化文件，请运行 make fmt" >&2; exit 1; }
	@out=$$($(GOIMPORTS) -l .); test -z "$$out" || { echo "$$out"; echo "存在未整理 import，请运行 make fmt" >&2; exit 1; }

vet: ## go vet
	@pkgs="$$($(GO) list ./... 2>/dev/null)"; if [ -n "$$pkgs" ]; then $(GO) vet $$pkgs; else echo "暂无 Go 包，跳过 vet"; fi

golangci-lint: ## golangci-lint
	@pkgs="$$($(GO) list ./... 2>/dev/null)"; if [ -n "$$pkgs" ]; then $(GOLANGCI) run; else echo "暂无 Go 包，跳过 golangci-lint"; fi

lint: fmt-check vet golangci-lint ## 格式 + vet + 静态检查

test: ## 单元测试
	@pkgs="$$($(GO) list ./... 2>/dev/null)"; if [ -n "$$pkgs" ]; then $(GO) test $$pkgs; else echo "暂无 Go 包，跳过 test"; fi

build: ## 编译全部包
	$(GO) build ./...

tidy: ## 整理依赖
	$(GO) mod tidy

seed: ## 造测试数据（里程碑 A 后续实现）
	@echo "目标 '$@' 尚未实现（里程碑 A 后续）" >&2

warmup: ## 预热活动库存到 Redis（里程碑 B）
	@echo "目标 '$@' 尚未实现（里程碑 B）" >&2

run: ## 本地运行服务（里程碑 A 后续实现）
	@echo "目标 '$@' 尚未实现（里程碑 A 后续）" >&2

loadtest: ## k6 压测（里程碑 C）
	@echo "目标 '$@' 尚未实现（里程碑 C）" >&2
