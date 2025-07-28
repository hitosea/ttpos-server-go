LOCAL_IP := $(shell ifconfig | grep "inet " | grep "192" | awk '{print $$2}' | head -n 1)

# 定义一个函数来更新环境变量并执行脚本
define update_env_and_debug
	sed -i.bak 's/^SERVER_MODE=.*/SERVER_MODE=debug/' .env && rm .env.bak;
	sed -i.bak 's/^APP_DEBUG=.*/APP_DEBUG=true/' .env && rm .env.bak;
endef

# 定义一个函数来更新环境变量并执行脚本
define update_env_and_run
	sed -i.bak 's/^DB_HOST=.*/DB_HOST=$(LOCAL_IP)/' .env && rm .env.bak;
	sed -i.bak 's/^REDIS_HOST=.*/REDIS_HOST=$(LOCAL_IP),$(LOCAL_IP),$(LOCAL_IP)/' .env && rm .env.bak;
	sed -i.bak 's/^REDIS_PORT=.*/REDIS_PORT=7001,7002,7003/' .env && rm .env.bak;
	if grep -q '^REDIS_CLUSTER_ANNOUNCE_IP=' .env; then \
		sed -i.bak 's/^REDIS_CLUSTER_ANNOUNCE_IP=.*/REDIS_CLUSTER_ANNOUNCE_IP=$(LOCAL_IP)/' .env && rm .env.bak; \
	else \
		echo '\nREDIS_CLUSTER_ANNOUNCE_IP=$(LOCAL_IP)' >> .env; \
	fi;
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh mysql open
endef

# 初始化项目
install:
	# 初始化env文件 
	if [ ! -f ".env" ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		sed -i.bak 's/^APP_ID=.*/APP_ID='$$(openssl rand -hex 3)'/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_PASSWORD=.*/DB_PASSWORD='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_ROOT_PASSWORD=.*/DB_ROOT_PASSWORD='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
	fi 
	# 检查 .env 文件是否存在
	if [ -f ".env" ]; then \
		sed -i.bak 's/^DB_HOST=.*/DB_HOST=db/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_PORT=.*/DB_PORT=3306/' .env && rm .env.bak; \
		sed -i.bak 's/^REDIS_HOST=.*/REDIS_HOST=redis/' .env && rm .env.bak; \
	fi
	# 启动容器
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d --build
    # 初始化php项目
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh init

# 重新构建项目
build:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d --build
	make migrate

build-run:
	make build

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
	$(call update_env_and_run)

# 运行数据库迁移
migrate:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh think migrate:run
	#更新 takeout模块数据库
	cd takeout && make conf && make db_up.docker

# 生成文档
build-doc:
	cd main && go install github.com/swaggo/swag/cmd/swag@latest && ${HOME}/go/bin/swag init

# 重启容器
restart:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh restart $(filter-out $@,$(MAKECMDGOALS))

# 更新redis
up:
	chmod +x ./scripts/cmd.sh && ./scripts/cmd.sh up -d

# 翻译
translate:
	cd main && go run ./main.go translate

# 运行数据库迁移
migrate-data:
	cd main && go run ./main.go migrate-data

# 快速增加版本号
add-version:
	@echo "快速增加版本号..."
	@CURRENT_VERSION=$$(grep 'Version.*=.*"' main/version/version.go | sed 's/.*"\(.*\)".*/\1/'); \
	MAJOR=$$(echo $$CURRENT_VERSION | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT_VERSION | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT_VERSION | cut -d. -f3); \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_VERSION="$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	CURRENT_DATE=$$(date +%Y-%m-%d); \
	CURRENT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	echo "当前版本: $$CURRENT_VERSION, 新版本: $$NEW_VERSION"; \
	cd main && go run ./main.go version --version=$$NEW_VERSION --commit=$$CURRENT_COMMIT --build-time=$$CURRENT_DATE

# 统计数据重跑
statistics-re:
	cd main && go run ./main.go statistics-re $(ARGS)

# 更新skootar状态
skootar-update-status:
	cd main && go run ./main.go skootar-update-status $(ARGS)

# 忽略不存在的目标（用于处理额外参数）
.PHONY: $(filter-out $(firstword $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(filter-out $(firstword $(MAKECMDGOALS)),$(MAKECMDGOALS)):
	@: