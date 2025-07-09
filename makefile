LOCAL_IP := $(shell ifconfig | grep "inet " | grep "192" | awk '{print $$2}' | head -n 1)

# 定义一个函数来更新环境变量并执行脚本
define update_env_and_debug
	sed -i.bak 's/^SERVER_MODE=.*/SERVER_MODE=debug/' .env && rm .env.bak;
	sed -i.bak 's/^APP_DEBUG=.*/APP_DEBUG=true/' .env && rm .env.bak;
endef

# 定义一个函数来更新环境变量并执行脚本
define update_env_and_run
	sed -i.bak 's/^DB_HOST=.*/DB_HOST=$(LOCAL_IP)/' .env && rm .env.bak;
	sed -i.bak 's/^REDIS_HOST=.*/REDIS_HOST=$(LOCAL_IP)/' .env && rm .env.bak;
	chmod +x ./.sh && ./.sh mysql open
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
		# 新增中台配置 \
		sed -i.bak 's/^NACOS_AUTH_TOKEN=.*/NACOS_AUTH_TOKEN='$$(openssl rand -hex 32 | base64)'/' .env && rm .env.bak; \
        sed -i.bak 's/^NACOS_AUTH_IDENTITY_KEY=.*/NACOS_AUTH_IDENTITY_KEY='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
		sed -i.bak 's/^NACOS_AUTH_IDENTITY_VALUE=.*/NACOS_AUTH_IDENTITY_VALUE='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
	fi 
	# 检查 .env 文件是否存在
	if [ -f ".env" ]; then \
		sed -i.bak 's/^DB_HOST=.*/DB_HOST=db/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_PORT=.*/DB_PORT=3306/' .env && rm .env.bak; \
		sed -i.bak 's/^REDIS_HOST=.*/REDIS_HOST=redis/' .env && rm .env.bak; \
	fi
	# 启动容器
	docker compose -p ttpos-server-go up -d --build;
    # 初始化php项目
	chmod +x ./.sh && ./.sh init

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
	chmod +x ./.sh && ./.sh think migrate:run

# 生成文档
build-doc:
	cd main && go install github.com/swaggo/swag/cmd/swag@latest && ${HOME}/go/bin/swag init

# 构建项目 - 生产
build-run:
	docker compose -p ttpos-server-go up -d --build

# 更新
update:
	chmod +x ./.sh && ./.sh update
	docker compose -p ttpos-server-go up -d --build
	chmod +x ./.sh && ./.sh restart

# 重启容器
restart:
	chmod +x ./.sh && ./.sh restart

# 翻译
translate:
	cd main && go run ./main.go translate

# 运行数据库迁移
migrate-data:
	cd main && go run ./main.go migrate-data $(ARGS)

# 更新版本号
update-version:
	@echo "更新版本号..."
	@CURRENT_COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	if [ -z "$(VERSION)" ] && [ -z "$(COMMIT_SHA)" ] && [ -z "$(BUILD_TIME)" ]; then \
		echo "请提供至少一个参数: VERSION 或 BUILD_TIME"; \
		echo "示例: make update-version VERSION=2.3.1 BUILD_TIME=2023-10-15"; \
		echo "注意: COMMIT_SHA 将自动使用当前最新的Git提交哈希值"; \
		echo "快速更新版本: make bump-version"; \
		cd main && go run ./main.go version; \
	else \
		cd main && go run ./main.go version \
		$(if $(VERSION),--version=$(VERSION),) \
		--commit=$$CURRENT_COMMIT \
		$(if $(BUILD_TIME),--build-time=$(BUILD_TIME),); \
	fi

# 快速增加版本号
bump-version:
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

