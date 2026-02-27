.PHONY: help build run dev clean test fmt vet tidy docker-build docker-run

# 默认目标
help:
	@echo "ttpos-ai Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        构建二进制文件"
	@echo "  make run          运行服务"
	@echo "  make dev          开发模式运行（热重载需安装 air）"
	@echo "  make test         运行测试"
	@echo "  make fmt          格式化代码"
	@echo "  make vet          静态检查"
	@echo "  make tidy         整理依赖"
	@echo "  make clean        清理构建产物"
	@echo "  make docker-build 构建 Docker 镜像"
	@echo "  make docker-run   运行 Docker 容器"
	@echo ""

# 变量
APP_NAME := ttpos-ai
BUILD_DIR := ./build
MAIN_FILE := ./cmd/main.go
DOCKER_IMAGE := ttpos-ai:latest

# 构建
build:
	@echo ">>> 构建 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo ">>> 构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run:
	@echo ">>> 运行 $(APP_NAME)..."
	go run $(MAIN_FILE)

# 开发模式（需要安装 air: go install github.com/air-verse/air@latest）
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air 未安装，使用普通模式运行"; \
		go run $(MAIN_FILE); \
	fi

# 测试
test:
	@echo ">>> 运行测试..."
	go test -v ./...

# 测试覆盖率
test-cover:
	@echo ">>> 运行测试覆盖率..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo ">>> 覆盖率报告: coverage.html"

# 格式化
fmt:
	@echo ">>> 格式化代码..."
	go fmt ./...

# 静态检查
vet:
	@echo ">>> 静态检查..."
	go vet ./...

# 整理依赖
tidy:
	@echo ">>> 整理依赖..."
	go mod tidy

# 清理
clean:
	@echo ">>> 清理构建产物..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo ">>> 清理完成"

# Docker 构建
docker-build:
	@echo ">>> 构建 Docker 镜像..."
	docker build -t $(DOCKER_IMAGE) .
	@echo ">>> 镜像构建完成: $(DOCKER_IMAGE)"

# Docker 运行
docker-run:
	@echo ">>> 运行 Docker 容器..."
	docker run --rm -p 8090:8090 --env-file ../.env $(DOCKER_IMAGE)

# 生成 Swagger 文档（需要安装 swag）
swagger:
	@if command -v swag > /dev/null; then \
		swag init -g cmd/main.go -o docs; \
	else \
		echo "swag 未安装，请运行: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# 一键检查（格式化 + 静态检查 + 测试）
check: fmt vet test
	@echo ">>> 检查完成"
