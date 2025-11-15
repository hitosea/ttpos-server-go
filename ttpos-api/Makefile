# TTPOS API Makefile

.PHONY: help tidy fmt vet test clean

# 默认目标
help:
	@echo "TTPOS API 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  make tidy    - 整理依赖"
	@echo "  make fmt     - 格式化代码"
	@echo "  make vet     - 静态检查"
	@echo "  make test    - 运行测试"
	@echo "  make clean   - 清理临时文件"
	@echo "  make all     - 执行所有检查"

# 整理依赖
tidy:
	@echo "整理依赖..."
	go mod tidy
	@echo "依赖整理完成"

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "代码格式化完成"

# 静态检查
vet:
	@echo "静态检查..."
	go vet ./...
	@echo "静态检查完成"

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...
	@echo "测试完成"

# 清理临时文件
clean:
	@echo "清理临时文件..."
	go clean
	@echo "清理完成"

# 执行所有检查
all: tidy fmt vet test
	@echo "所有检查完成"

