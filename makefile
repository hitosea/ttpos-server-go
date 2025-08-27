# include
include ./scripts/cmd.mk

# 初始化项目
install:
	@make init-env
	@make build-web
	@make redis-clear-data-node-conf
	# 启动容器
	@echo "🗄️  启动容器..."
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d --build
	@echo "🗄️  初始化php项目..."
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh init
	@$(call update_env_and_run)
	@make install-bmp
	@echo "✅ 初始化完成"

# 初始化中台模块
install-bmp:
	@echo "🗄️  初始化中台模块..."
	@make init-bmp-env
	@cd ttpos-bmp && make conf && make migrate && make mid && make up

# 重新构建项目
build:
	@make build-web
	@make redis-clear-data-node-conf
	@echo "🐳 构建 Docker 容器..."
	@cd ./main && GOOS=linux GOARCH=amd64 go build -o main main.go
	@chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d --build
	@echo "✅ Docker 构建完成"
	@echo "🗄️  运行数据库迁移..."
	@make migrate
	@echo "✅ 构建完成"

# 生成文档
build-doc:
	cd main && go install github.com/swaggo/swag/cmd/swag@latest && ${HOME}/go/bin/swag init

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

# 开启mysql端口
mysql-open:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh mysql open

# 运行数据库迁移
migrate:
	@make check-db-host-open-mysql
	@echo "🗄️  运行主项目数据库迁移..."
	@chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh think migrate:run
	@echo "🚀 更新 中台 模块数据库..."
	@cd ttpos-bmp && make conf && make migrate
	@echo "✅ 数据库迁移完成"


# 重启容器
restart:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh restart $(filter-out $@,$(MAKECMDGOALS))

# docker-compose up -d
up:
	@make redis-clear-data-node-conf > /dev/null 2>&1
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d
	@echo "🔍 启动HTTP调试代理..."
	@make start-http-debug-proxy

# docker-compose ps
ps:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh ps

# docker-compose down
down:
	@make stop-http-debug-proxy
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh down $(filter-out $@,$(MAKECMDGOALS))
	
# 翻译
translate:
	cd main && go run ./main.go translate

# 运行数据库旧数据迁移
migrate-data:
	cd main && go run ./main.go migrate-data

# 统计数据重跑
statistics-re:
	cd main && go run ./main.go statistics-re $(ARGS)

# 更新skootar状态
skootar-update-status:
	cd main && go run ./main.go skootar-update-status $(ARGS)

# 重置密码
repassword:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh repassword $(ARGS)

# 整理依赖
mod-tidy:
	@cd main && go mod tidy

# 增加版本号
add-ver:
	@make add-version

# 执行think命令
think:
	@chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh think $(filter-out $@,$(MAKECMDGOALS)) 