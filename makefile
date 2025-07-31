# include
include ./scripts/cmd.mk

# 初始化项目
install:
	make init-env
	make build-web
	make redis-clear-data
	# 启动容器
	@echo "🗄️  启动容器..."
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d --build
	@echo "🗄️  初始化php项目..."
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh init
	@echo "🗄️  初始化takeout模块..."
	cd takeout && make conf && make db_up.docker
	@echo "✅ 初始化完成"

# 重新构建项目
build:
	make build-web
	make redis-clear-data
	@echo "🐳 构建 Docker 容器..."
	@chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d --build
	@echo "✅ Docker 构建完成"
	@echo "🗄️  运行数据库迁移..."
	@make migrate
	@echo "✅ 构建完成"

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
	if [ ! -f "${HOME}/go/bin/fresh" ]; then \
		go install github.com/pilu/fresh@latest; \
	fi
	cd main && ${HOME}/go/bin/fresh -c ./fresh.conf

# 开启mysql端口
mysql-open:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh mysql open

# 运行数据库迁移
migrate:
	make check-db-host-open-mysql
	@echo "🗄️  运行主项目数据库迁移..."
	@chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh think migrate:run
	@echo "🚀 更新 takeout 模块数据库..."
	@cd takeout && make conf && make db_up.docker
	@echo "✅ 数据库迁移完成"

# 生成文档
build-doc:
	cd main && go install github.com/swaggo/swag/cmd/swag@latest && ${HOME}/go/bin/swag init

# 重启容器
restart:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh restart $(filter-out $@,$(MAKECMDGOALS))

# docker-compose up -d
up:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d

# docker-compose ps
ps:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh ps

# docker-compose down
down:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh down
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

# 增加版本号
add-ver:
	make add-version