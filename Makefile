# include
include ./ttpos-scripts/cmd.mk

# 初始化项目
install:
	@make init-env
	@make build-web
	@make redis-clear-data-node-conf
	# 启动容器
	@echo "🗄️  启动容器..."
	chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh up -d --build
	@echo "🗄️  初始化php项目..."
	chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh init
	@$(call update_env_and_run)
	@make bmp-install
	@echo "✅ 初始化完成"

# 初始化中台模块
bmp-install:
	@echo "🗄️  初始化中台模块..."
	@make bmp-init-env
	@cd ttpos-bmp && make update-ip && make conf && make mid && make migrate && make up

# 重新构建项目
build:
	@make build-web
	@make redis-clear-data-node-conf
	@echo "🐳 构建 Docker 容器..."
	@if command -v go >/dev/null 2>&1; then \
		echo "🔨 在宿主机上构建 Go 二进制文件..."; \
		cd ./main && GOOS=linux GOARCH=amd64 go build -o main main.go; \
	else \
		echo "⚠️  宿主机未安装 Go，将跳过本地构建，使用 Docker 构建..."; \
	fi
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh up -d --build
	@echo "✅ Docker 构建完成"
	@$(call update_env_and_run)
	@echo "🗄️  运行数据库迁移..."
	@make migrate
	@make rmi-docker-images
	@echo "✅ 构建完成"

#重新构建中台模块
build-bmp:
	@echo "🐳 构建中台模块..."
	@cd ttpos-bmp && make up

# 生成文档
build-doc:
	cd main && go install github.com/swaggo/swag/cmd/swag@latest && ${GO_PATH}/bin/swag init

# 变更debug模式
debug:
	$(call update_env_and_debug)

# 运行run
run:
	@make stop-service
	cd main && go run ./main.go

# 启动开发模式 - 热重启
dev: debug
	$(call update_env_and_run)
	if [ ! -f "$(GO_PATH)/bin/fresh" ]; then \
		go install github.com/pilu/fresh@latest; \
	fi
	cd main && $(GO_PATH)/bin/fresh -c ./fresh.conf

# 运行数据库迁移
migrate:
	@echo "🗄️  运行主项目数据库迁移..."
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh php-migrate
	@echo "🚀 更新 中台 模块数据库..."
	@cd ttpos-bmp && make update-ip && make conf && make migrate
	@echo "✅ 数据库迁移完成"


# 重启容器
restart:
	chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh restart $(filter-out $@,$(MAKECMDGOALS))

# docker-compose up -d
up:
	@make redis-clear-data-node-conf
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh up -d
	@echo "🔍 启动HTTP调试代理..."
	@make start-http-debug-proxy

# up 中台模块
up-bmp:
	@cd ttpos-bmp && make up

# docker-compose ps
ps:
	chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh ps

# docker-compose down
down:
	@make stop-http-debug-proxy
	chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh down $(filter-out $@,$(MAKECMDGOALS))
	
# 翻译
translate:
	cd main && go run ./main.go translate

# 运行数据库旧数据迁移admin
migrate-data:
	cd main && go run ./main.go migrate-data

# 重新同步本地产品数据ERP
resync-product-data-to-erp:
	cd main && go run ./main.go resync-product-data-to-erp

# 同步ERP数据
sync-erp-data:
	cd main && go run ./main.go sync-erp-data

# 统计数据重跑
statistics-re:
	cd main && go run ./main.go statistics-re $(ARGS)

# 更新skootar状态
skootar-update-status:
	cd main && go run ./main.go skootar-update-status $(ARGS)

# 重置密码
repassword:
	chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh repassword $(ARGS)

# 整理依赖
mod-tidy:
	@cd main && go mod tidy

# 执行think命令
think:
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh think $(filter-out $@,$(MAKECMDGOALS)) 

# 监听今天的日志（格式化输出，显示完整JSON）
log:
	@echo "🔍 监听今天的日志（完整JSON格式，忽略warn级别）..."
	@tail -f -n 100 ./main/log/$$(date +%Y-%m-%d).log | grep --line-buffered -v '"level":"warn"' | grep --line-buffered -v '"level":"debug"' | while IFS= read -r line; do \
		if echo "$$line" | grep -q '"level":"error"'; then \
			echo "$$line" | jq -C . 2>/dev/null || echo "$$line" | sed 's/.*/\x1b[0;31m&\x1b[0m/'; \
		else \
			echo "$$line" | jq -C . 2>/dev/null || echo "$$line"; \
		fi; \
	done

# 添加物品库存
add-item-stock:
	cd main && go run ./main.go add-item-stock

# 添加父级公司UUID路径
add-parent-company-uuid:
	cd main && go run ./main.go add-parent-company-uuid

# 删除 dangling 镜像
rmi-docker-images:
	docker rmi $$(docker images -qf "dangling=true")

# 授权所有文件
chown-all:
	sudo chown -R $(shell whoami):$(shell id -gn) $(CURDIR)
	sudo chown -R www-data:www-data $(CURDIR)/admin/runtime
	sudo chown -R www-data:www-data $(CURDIR)/admin/public/uploads
	sudo chmod -R 755 $(CURDIR)/admin/runtime
	sudo rm -rf $(CURDIR)/admin/runtime/logs/*

# 更新 MCP Token
update-mcp-token:
	@echo "🔄 更新 MCP Token..."
	bash ./ttpos-scripts/dootask.sh

# 验证数据库结构
verify-db-structure:
	cd ./main && go run main.go verify-db-structure

# 导出数据库备份
db-export:
	@echo "🗄️  导出数据库备份..."
	@APP_ID=$$(grep '^APP_ID=' .env | cut -d '=' -f 2 | tr -d ' '); \
	DB_ROOT_PASSWORD=$$(grep '^DB_ROOT_PASSWORD=' .env | cut -d '=' -f 2 | tr -d ' '); \
	CONTAINER="saas-db-$$APP_ID"; \
	BACKUP_FILE="backup_all_$$(date +%Y%m%d_%H%M%S).sql.gz"; \
	echo "📦 容器: $$CONTAINER"; \
	echo "📁 备份文件: $$BACKUP_FILE"; \
	docker exec $$CONTAINER mysqldump -uroot -p$$DB_ROOT_PASSWORD \
		--all-databases \
		--single-transaction \
		--routines \
		--triggers \
		--events \
		--set-gtid-purged=OFF \
		| gzip > $$BACKUP_FILE; \
	echo "✅ 数据库备份完成: $$BACKUP_FILE ($$(du -h $$BACKUP_FILE | cut -f1))"

# 导入数据库备份
# 用法: make db-import FILE=backup_all_20260307.sql.gz
db-import:
	@if [ -z "$(FILE)" ]; then \
		echo "❌ 请指定备份文件: make db-import FILE=backup_all_xxx.sql.gz"; \
		exit 1; \
	fi
	@if [ ! -f "$(FILE)" ]; then \
		echo "❌ 文件不存在: $(FILE)"; \
		exit 1; \
	fi
	@echo "🗄️  导入数据库备份: $(FILE)"
	@APP_ID=$$(grep '^APP_ID=' .env | cut -d '=' -f 2 | tr -d ' '); \
	DB_ROOT_PASSWORD=$$(grep '^DB_ROOT_PASSWORD=' .env | cut -d '=' -f 2 | tr -d ' '); \
	CONTAINER="saas-db-$$APP_ID"; \
	echo "📦 容器: $$CONTAINER"; \
	echo "⏳ 正在导入，请稍候..."; \
	gunzip < $(FILE) | docker exec -i $$CONTAINER mysql -uroot -p$$DB_ROOT_PASSWORD; \
	echo "✅ 数据库导入完成"

# ===========================================
# Integration Testing Commands
# ===========================================

# Build configuration for tests
BUILD_ID ?= $(shell date +%s)
IMAGE_TAG ?= latest
COVERAGE_DIR := coverage
ENV_FILE := main/tests/test.env

test-help:
	@echo "TTPOS Integration Test Commands"
	@echo ""
	@echo "  make test-integration   Run integration tests (CI mode)"
	@echo "  make test-local         Run integration tests (local dev mode)"
	@echo "  make test-coverage      Alias for test-local (backward compatibility)"
	@echo "  make test-clean         Clean up test containers and volumes"
	@echo ""
	@echo "Coverage is always collected and merged after tests complete."
	@echo ""
	@echo "Options:"
	@echo "  BUILD_ID=<id>           Unique identifier for test run (default: timestamp)"

# Run integration tests (CI mode)
test-integration:
	@echo "=== Running Integration Tests (CI Mode) ==="
	@echo "Build ID: test-$(BUILD_ID)"
	@mkdir -p $(COVERAGE_DIR)
	docker compose --env-file $(ENV_FILE) -p test-$(BUILD_ID) -f main/tests/docker-compose.yml down -v --remove-orphans 2>/dev/null || true
	docker compose --env-file $(ENV_FILE) -p test-$(BUILD_ID) -f main/tests/docker-compose.yml up --build --exit-code-from test-runner
	@echo "=== Merging coverage data ==="
	go tool covdata textfmt -i=$(COVERAGE_DIR) -o=$(COVERAGE_DIR)/total.out 2>/dev/null || echo "No coverage data to merge"
	@if [ -f $(COVERAGE_DIR)/total.out ]; then bash main/tests/fix-coverage-paths.sh $(COVERAGE_DIR)/total.out; fi
	@echo "Coverage report: $(COVERAGE_DIR)/total.out"
	@echo "=== Run 'make test-clean' to clean up ==="

# Run integration tests locally (builds from source)
test-local:
	@echo "=== Running Integration Tests (Local Mode) ==="
	@echo "=== Cleaning up any previous run ==="
	docker compose --env-file $(ENV_FILE) -p test-$(BUILD_ID) -f main/tests/docker-compose.yml down -v --remove-orphans 2>/dev/null || true
	@echo "=== Starting containers ==="
	@mkdir -p $(COVERAGE_DIR)
	docker compose --env-file $(ENV_FILE) -p test-$(BUILD_ID) -f main/tests/docker-compose.yml up --build --exit-code-from test-runner
	@echo "=== Merging coverage data ==="
	go tool covdata textfmt -i=$(COVERAGE_DIR) -o=$(COVERAGE_DIR)/total.out 2>/dev/null || echo "No coverage data to merge"
	@if [ -f $(COVERAGE_DIR)/total.out ]; then bash main/tests/fix-coverage-paths.sh $(COVERAGE_DIR)/total.out; fi
	@echo "Coverage report: $(COVERAGE_DIR)/total.out"
	@echo "=== Run 'make test-clean' to clean up ==="

# Alias for test-local (backward compatibility)
test-coverage: test-local

# Clean up all test containers and volumes
test-clean:
	@echo "=== Cleaning up all test containers ==="
	@docker compose --env-file $(ENV_FILE) -p test-$(BUILD_ID) -f main/tests/docker-compose.yml down -v --remove-orphans 2>/dev/null || true
	@docker ps -q --filter "name=test-*" | xargs -r docker rm -f 2>/dev/null || true
	@echo "=== Cleaning up coverage directory ==="
	@rm -rf $(COVERAGE_DIR)/*
	@echo "Done."

# ===========================================
# Unit Testing Commands
# ===========================================

.PHONY: unit-test unit-test-coverage

# Run unit tests
unit-test:
	@echo "=== Running Unit Tests ==="
	cd main && go test -v -count=1 ./...

# Run unit tests with coverage
unit-test-coverage:
	@echo "=== Running Unit Tests with Coverage ==="
	@mkdir -p $(COVERAGE_DIR)
	cd main && go test -coverprofile=../$(COVERAGE_DIR)/unit.out -covermode=atomic ./...

# ===========================================
# Code Quality Commands
# ===========================================

.PHONY: lint fmt vet sonar

# Run linter
lint:
	@echo "=== Running Linter ==="
	cd main && golangci-lint run ./...

# Format code
fmt:
	@echo "=== Formatting Code ==="
	cd main && go fmt ./...

# Run go vet
vet:
	@echo "=== Running go vet ==="
	cd main && go vet ./...
