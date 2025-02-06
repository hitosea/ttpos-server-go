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
	fi
	# 启动容器
	docker compose -p ttpos-server-go up -d;
    # 初始化php项目
	chmod +x ./.sh && ./.sh init

# 变更debug模式
debug:
	$(call update_env_and_debug)

# 运行run
run: debug
	$(call update_env_and_run)
	cd main && go run ./cmd/server/main.go

# 启动开发模式 - 热重启
dev: debug
	$(call update_env_and_run)
	if [ ! -f "${HOME}/go/bin/fresh" ]; then \
		go install github.com/pilu/fresh@latest; \
	fi
	cd main && ${HOME}/go/bin/fresh -c ./fresh.conf

# 运行数据库迁移
migrate:
	chmod +x ./.sh && ./.sh think migrate:run

# 生成文档
build-doc:
	cd main && go install github.com/swaggo/swag/cmd/swag@latest && chmod +x ./scripts/build.sh && ./scripts/build.sh swagger

# 构建项目 - 生产
build-run:
	chmod +x ./.sh && ./.sh golang go build -o main ./cmd/server/main.go
	docker compose -p ttpos-server-go restart golang

# 更新
update:
	chmod +x ./.sh && ./.sh update
	chmod +x ./.sh && ./.sh golang go build -o main ./cmd/server/main.go
	docker compose -p ttpos-server-go restart golang