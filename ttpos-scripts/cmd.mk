LOCAL_IP := $(shell ifconfig | grep "inet " | grep "192" | awk '{print $$2}' | head -n 1)
ifeq ($(LOCAL_IP),)
	LOCAL_IP := $(shell ttpos-scripts/get_ip.sh)
endif
GO_PATH := $(shell go env GOPATH)

# 定义一个函数来更新环境变量并执行脚本
define update_env_and_debug
	sed -i.bak 's/^SERVER_MODE=.*/SERVER_MODE=debug/' .env && rm .env.bak;
	sed -i.bak 's/^APP_DEBUG=.*/APP_DEBUG=true/' .env && rm .env.bak;
endef

# 定义一个函数来更新环境变量并执行脚本
define update_env_and_run
	sed -i.bak 's/^DB_HOST=.*/DB_HOST=$(LOCAL_IP)/' .env && rm .env.bak;
	DB_PORT_VALUE=$$(grep '^DB_PORT_OPEN=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' ' || echo ""); \
	if [ "$$DB_PORT_VALUE" != "" ]; then \
		sed -i.bak "s/^DB_PORT=.*/DB_PORT=$$DB_PORT_VALUE/" .env && rm .env.bak; \
	fi;
	sed -i.bak 's/^REDIS_HOST=.*/REDIS_HOST=$(LOCAL_IP),$(LOCAL_IP),$(LOCAL_IP)/' .env && rm .env.bak;
	sed -i.bak 's/^REDIS_PORT=.*/REDIS_PORT=7001,7002,7003/' .env && rm .env.bak;
	if grep -q '^REDIS_CLUSTER_ANNOUNCE_IP=' .env; then \
		sed -i.bak 's/^REDIS_CLUSTER_ANNOUNCE_IP=.*/REDIS_CLUSTER_ANNOUNCE_IP=$(LOCAL_IP)/' .env && rm .env.bak; \
	else \
		echo '\nREDIS_CLUSTER_ANNOUNCE_IP=$(LOCAL_IP)' >> .env; \
	fi;
	chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh up -d
	@make start-http-debug-proxy
endef

# 默认目标 - 显示帮助信息
.DEFAULT_GOAL := help

# 显示帮助信息
help:
	@printf "\033[1;36m"
	@echo " 🍽️  TTPOS 餐饮收银系统 - 命令手册 🍽️                     "
	@printf "\033[0m"
	@echo ""
	@printf "\033[1;32m"
	@echo "🚀 项目管理命令"
	@printf "\033[0m"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "install" "初始化项目（首次安装）"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "build" "重新构建项目"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "init-env" "初始化env文件"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "restart" "重启容器"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "up" "启动Docker容器"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "down" "停止Docker容器"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "ps" "查看容器状态"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "mod-tidy" "整理依赖"
	@echo ""
	@echo "🚀 中台管理命令"
	@printf "\033[0m"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "bmp-init-env" "初始化中台env文件"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "bmp-install" "初始化并启动中台模块"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "bmp-conf" "配置中台模块"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "bmp-migrate" "迁移中台模块"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "bmp-up" "启动中台模块"
	@echo ""

	@printf "\033[1;32m"
	@echo "🔧 开发命令"
	@printf "\033[0m"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "debug" "切换到调试模式"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "run" "运行项目（调试模式）"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "dev" "启动开发模式（热重启 + HTTP调试代理）"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "build-web" "构建前端项目"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "build-doc" "生成API文档"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "start-http-debug-proxy" "启动HTTP调试代理"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "stop-http-debug-proxy" "停止HTTP调试代理"
	@echo ""
	@printf "\033[1;32m"
	@echo "🗄️  数据库命令"
	@printf "\033[0m"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "migrate" "运行数据库迁移"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "migrate-data" "运行旧数据迁移"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "mysql-open" "开启MySQL端口"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "sync-erp-data" "同步ERP数据"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "resync-product-data-to-erp" "重新同步本地产品数据ERP"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "check-db-host-open-mysql" "检查DB_HOST并开启MySQL端口"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "add-item-stock" "添加物品库存"
	@echo ""
	@printf "\033[1;32m"
	@echo "🔐 系统管理命令"
	@printf "\033[0m"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "repassword" "重置密码"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "translate" "运行翻译命令"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "statistics-re" "重新统计数据"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "skootar-update-status" "更新Skootar状态"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "think" "执行think命令"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "chown-all" "修改文件权限"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "rmi-all" "删除所有无用镜像"
	@echo ""
	@printf "\033[1;32m"
	@echo "📦 版本管理"
	@printf "\033[0m"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "add-ver" "增加版本号"
	@echo ""
	@printf "\033[1;32m"
	@echo "🧹 清理命令"
	@printf "\033[0m"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "redis-clear-data-node-conf" "清空Redis集群配置数据"
	@echo ""
	@printf "\033[1;32m"
	@echo "🔍 日志命令"
	@printf "\033[1;33m  %-25s\033[0m - %s\n" "log" "监听今天的日志"
	@echo ""
	@printf "\033[1;35m"
	@echo ""

# 初始化项目
init-env:
	@echo "🔍 初始化env文件"
	@corepack enable
	@if [ ! -f ".env" ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		sed -i.bak 's/^APP_ID=.*/APP_ID='$$(openssl rand -hex 3)'/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_PASSWORD=.*/DB_PASSWORD='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_ROOT_PASSWORD=.*/DB_ROOT_PASSWORD='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
	fi
	@echo "🔍 检查 .env 文件是否存在"
	$(call update_env_and_run);

bmp-init-env:
	@echo "🔍 初始化中台env文件"
	@if [ ! -f "ttpos-bmp/.env" ]; then \
		cp ttpos-bmp/.env.example ttpos-bmp/.env; \
		echo "Created ttpos-bmp/.env file from ttpos-bmp/.env.example"; \
		sed -i.bak 's/^APP_ID=.*/APP_ID='$$(openssl rand -hex 3)'/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^NACOS_AUTH_TOKEN=.*/NACOS_AUTH_TOKEN='$$(openssl rand -hex 32 | base64)'/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^NACOS_AUTH_IDENTITY_KEY=.*/NACOS_AUTH_IDENTITY_KEY='$$(openssl rand -hex 8)'/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^NACOS_AUTH_IDENTITY_VALUE=.*/NACOS_AUTH_IDENTITY_VALUE='$$(openssl rand -hex 8)'/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^NACOS_SERVER_IP=.*/NACOS_SERVER_IP=10.0.11.10/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^ROCKETMQ_NAME_SRV_ADDR=.*/ROCKETMQ_NAME_SRV_ADDR=10.0.11.40:9876/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^ROCKETMQ_BROKER_ADDR=.*/ROCKETMQ_BROKER_ADDR=10.0.11.41:10911/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
	fi
	@echo "🔍 检查 ttpos-bmp/.env 文件是否存在"
	@if [ -f "ttpos-bmp/.env" ]; then \
		sed -i.bak 's/^DB_HOST=.*/DB_HOST=$(LOCAL_IP)/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		if [ -f ".env" ]; then \
			echo "🔄 从上级目录的 .env 文件同步数据库密码..."; \
			PARENT_DB_PASSWORD=$$(grep '^DB_PASSWORD=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' ' || echo ""); \
			PARENT_DB_ROOT_PASSWORD=$$(grep '^DB_ROOT_PASSWORD=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' ' || echo ""); \
			PARENT_DB_USERNAME=$$(grep '^DB_USERNAME=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' ' || echo ""); \
			if [ "$$PARENT_DB_PASSWORD" != "" ]; then \
				sed -i.bak "s/^DB_PASSWORD=.*/DB_PASSWORD=$$PARENT_DB_PASSWORD/" ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
				echo "✅ 已同步 DB_PASSWORD"; \
			fi; \
			if [ "$$PARENT_DB_ROOT_PASSWORD" != "" ]; then \
				sed -i.bak "s/^DB_ROOT_PASSWORD=.*/DB_ROOT_PASSWORD=$$PARENT_DB_ROOT_PASSWORD/" ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
				echo "✅ 已同步 DB_ROOT_PASSWORD"; \
			fi; \
			if [ "$$PARENT_DB_USERNAME" != "" ]; then \
				sed -i.bak "s/^DB_USERNAME=.*/DB_USERNAME=$$PARENT_DB_USERNAME/" ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
				echo "✅ 已同步 DB_USERNAME"; \
			fi; \
		fi; \
		sed -i.bak 's/^REDIS_HOST=.*/REDIS_HOST=$(LOCAL_IP),$(LOCAL_IP),$(LOCAL_IP)/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^REDIS_PORT=.*/REDIS_PORT=7001,7002,7003/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
		sed -i.bak 's/^GRPC_ENDPOINTS=.*/GRPC_ENDPOINTS=$(LOCAL_IP)/' ttpos-bmp/.env && rm ttpos-bmp/.env.bak; \
	fi

# 配置中台模块
bmp-conf:
	@make bmp-init-env
	@cd ttpos-bmp && make update-ip && make conf

# 迁移中台模块
bmp-migrate:
	@make bmp-init-env
	@cd ttpos-bmp && make update-ip && make conf && make migrate

# 启动中台模块
bmp-up:
	@make bmp-init-env
	@cd ttpos-bmp && make update-ip && make conf && make mid && make migrate && make up

# 构建前端
build-web:
	@echo "🔍 检查前端文件变化..."
	@if [ -d "admin/views" ]; then \
		FRONTEND_CHANGED=0; \
		BUILD_MARKER="./admin/runtime/frontend_build_marker"; \
		if [ ! -f "$$BUILD_MARKER" ]; then \
			echo "📝 首次构建，开始构建前端..."; \
			FRONTEND_CHANGED=1; \
		elif [ ! -d "admin/public/admin" ] || [ ! -d "admin/public/shop" ]; then \
			echo "📁 前端构建产物不存在，开始构建..."; \
			FRONTEND_CHANGED=1; \
		else \
			LAST_BUILD_COMMIT=$$(cat $$BUILD_MARKER 2>/dev/null || echo ""); \
			CURRENT_COMMIT=$$(git rev-parse HEAD 2>/dev/null || echo "current"); \
			if [ "$$LAST_BUILD_COMMIT" != "$$CURRENT_COMMIT" ]; then \
				if git diff --quiet $$LAST_BUILD_COMMIT..$$CURRENT_COMMIT -- admin/views/ 2>/dev/null; then \
					echo "✅ 自上次构建后前端文件无变化，跳过构建"; \
				else \
					echo "📝 检测到前端文件自上次构建后有变化，开始构建..."; \
					FRONTEND_CHANGED=1; \
				fi; \
			else \
				echo "✅ 前端文件无变化，跳过构建"; \
			fi; \
		fi; \
		if [ $$FRONTEND_CHANGED -eq 1 ]; then \
			echo "🚀 正在构建前端项目..."; \
			cd admin && ./build && echo "✅ 前端构建完成" || (echo "❌ 前端构建失败" && exit 1); \
			cd ..; \
			git rev-parse HEAD > $$BUILD_MARKER 2>/dev/null || echo "unknown" > $$BUILD_MARKER; \
		fi; \
	else \
		echo "⚠️  admin/views 目录不存在，跳过前端构建"; \
	fi

# 重新构建项目
build-run:
	make build

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

# 清空redis的cluster的data-*目录
redis-clear-data-node-conf:
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh down redis-node-1
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh down redis-node-2
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh down redis-node-3
	@sudo chown -R coder:coder /home/coder/workspaces/ttpos-server-go/docker > /dev/null 2>&1 || true
	@rm -rf ./docker/redis/cluster/data-* > /dev/null 2>&1 
	@sudo rm -rf ./docker/redis/cluster/data-* > /dev/null 2>&1 || true

# 检查env的DB_HOST是否等于 LOCAL_IP。等于的话 就执行 make mysql-open;
check-db-host-open-mysql:
	@if [ -f ".env" ]; then \
		DB_HOST_VALUE=$$(grep '^DB_HOST=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' '); \
		if [ "$$DB_HOST_VALUE" = "$(LOCAL_IP)" ]; then \
			echo "🔍 检测到DB_HOST为本地IP，开启MySQL端口..."; \
			make mysql-open; \
		else \
			echo "🔍 DB_HOST不是本地IP ($$DB_HOST_VALUE vs $(LOCAL_IP))，跳过开启MySQL端口"; \
		fi; \
	else \
		echo "⚠️  .env文件不存在，跳过MySQL端口检查"; \
	fi

# HTTP调试代理管理
start-http-debug-proxy:
	@if ! docker images | grep -q "602666178/http-proxy-debug-view"; then \
		echo "📥 镜像不存在，正在拉取..." && \
		docker pull 602666178/http-proxy-debug-view:latest; \
	else \
		echo "✅ 镜像已存在，跳过拉取"; \
	fi
	@NGINX_PORT=$$(grep '^NGINX_PORT=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' ' || echo "8888"); \
	SERVER_PORT=$$(grep '^DEBUG_SERVER_PORT=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' ' || echo ""); \
	if [ "$$SERVER_PORT" = "" ]; then \
		SERVER_PORT=$$(grep '^SERVER_PORT=' .env 2>/dev/null | cut -d '=' -f 2 | tr -d ' ' || echo "8080"); \
	fi; \
	PROXY_PORT=$$((SERVER_PORT + 1)); \
	TARGET_URL="http://$(LOCAL_IP):$$SERVER_PORT,http://$(LOCAL_IP):$$NGINX_PORT"; \
	echo "🎯 目标URL: $$TARGET_URL, 代理端口: $$PROXY_PORT"; \
	docker run -d \
		--name http-debug-proxy \
		-p $$PROXY_PORT:8080 \
		-p 8091:8091 \
		-e TARGET_URL="$$TARGET_URL" \
		-e PROXY_PORT=8080 \
		-e WEB_PORT=8091 \
		--restart unless-stopped \
		602666178/http-debug-proxy > /dev/null 2>&1 || \
		(echo "⚠️  HTTP调试代理容器已存在，正在重启..." && \
		docker rm -f http-debug-proxy > /dev/null 2>&1 && \
		docker run -d \
			--name http-debug-proxy \
			-p $$PROXY_PORT:8080 \
			-p 8091:8091 \
			-e TARGET_URL="$$TARGET_URL" \
			-e PROXY_PORT=8080 \
			-e WEB_PORT=8091 \
			--restart unless-stopped \
			602666178/http-debug-proxy:latest > /dev/null 2>&1); \
	echo "✅ HTTP调试代理已启动 - 代理端口: $$PROXY_PORT, 调试界面: http://localhost:8091"

stop-http-debug-proxy:
	@echo "🛑 停止HTTP调试代理容器..."
	@docker rm -f http-debug-proxy > /dev/null 2>&1 || echo "✅ HTTP调试代理已停止"

# 运行数据库迁移
php-migrate:
	@chmod +x ./ttpos-scripts/cmd.sh && ./ttpos-scripts/cmd.sh php-migrate

# 处理额外参数（不包含第一个目标）
ifneq ($(words $(MAKECMDGOALS)),1)
.PHONY: $(filter-out $(firstword $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(filter-out $(firstword $(MAKECMDGOALS)),$(MAKECMDGOALS)):
	@:
endif

# 捕获不存在的命令并显示帮助
%:
	@printf "\033[1;31mError:\033[0m make: *** No rule to make target '\033[1;33m$@\033[0m'. Stop.\n"
	@echo ""
	@printf "\033[1;37m🤔 您输入的命令 '\033[1;33m$@\033[1;37m' 不存在，请查看可用命令列表：\033[0m\n"
	@echo ""
	@make help