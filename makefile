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
run: debug
	$(call update_env_and_run)
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

# 增加版本号
add-ver:
	@make add-version

# 执行think命令
think:
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh think $(filter-out $@,$(MAKECMDGOALS)) 

# 监听今天的日志（格式化输出，显示完整JSON）
log:
	@echo "🔍 监听今天的日志（完整JSON格式）..."
	@tail -f -n 100 ./main/log/$$(date +%Y-%m-%d).log | while IFS= read -r line; do \
		echo "$$line" | jq . 2>/dev/null || echo "$$line"; \
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
	sudo chown -R coder:coder /home/coder/workspaces/ttpos-server-go

# 更新 MCP Token
update-mcp-token:
	bash /home/coder/workspaces/update_mcp_token.sh

# 验证数据库结构
verify-db-structure:
	cd ./main && go run main.go verify-db-structure