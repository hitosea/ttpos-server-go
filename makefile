# 初始化项目
install:
	# 初始化env文件 
	if [ ! -f ".env" ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		sed -i.bak 's/^APP_ID=.*/APP_ID='$$(openssl rand -hex 3)'/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_ROOT_PASSWORD=.*/DB_ROOT_PASSWORD='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
	fi
	# 启动容器
	docker compose -p ttpos-server-go up -d;
	# 运行run
	chmod +x ./.sh && ./.sh mysql open;
	cd main && go mod tidy;
    # 初始化php项目
	chmod +x ./.sh && ./.sh init
    # 启动容器
	docker compose -p ttpos-server-go restart;
    # 运行run
	cd main && go run ./cmd/server/main.go ./cmd/server/swagger_enabled.go;

# 运行run
run:
	chmod +x ./.sh && ./.sh mysql open
	cd main && go mod tidy
	cd main && go run ./cmd/server/main.go ./cmd/server/swagger_enabled.go

# 启动开发模式 - 热重启
dev:
	sed -i.bak 's/^SERVER_MODE=.*/SERVER_MODE=debug/' .env && rm .env.bak;
	chmod +x ./.sh && ./.sh mysql open
	if [ ! -f "${HOME}/go/bin/fresh" ]; then \
		go install github.com/pilu/fresh@latest; \
	fi
	cd main && ${HOME}/go/bin/fresh -c ./fresh.conf

migrate:
	cd main && go run ./migration/main.go run --version 1738765726

