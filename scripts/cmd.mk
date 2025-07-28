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
init-env:
	@echo "🔍 初始化env文件"
	if [ ! -f ".env" ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		sed -i.bak 's/^APP_ID=.*/APP_ID='$$(openssl rand -hex 3)'/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_PASSWORD=.*/DB_PASSWORD='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_ROOT_PASSWORD=.*/DB_ROOT_PASSWORD='$$(openssl rand -hex 8)'/' .env && rm .env.bak; \
	fi 
	@echo "🔍 检查 .env 文件是否存在"
	if [ -f ".env" ]; then \
		sed -i.bak 's/^DB_HOST=.*/DB_HOST=db/' .env && rm .env.bak; \
		sed -i.bak 's/^DB_PORT=.*/DB_PORT=3306/' .env && rm .env.bak; \
		sed -i.bak 's/^REDIS_HOST=.*/REDIS_HOST=redis/' .env && rm .env.bak; \
	fi

# 构建前端
build-web:
	@echo "🔍 检查前端文件变化..."
	@if [ -d "admin/views" ]; then \
		FRONTEND_CHANGED=0; \
		if git status --porcelain admin/views/ | grep -q .; then \
			echo "📝 检测到前端文件有变化，开始构建..."; \
			FRONTEND_CHANGED=1; \
		elif [ ! -d "admin/public/admin" ] || [ ! -d "admin/public/shop" ]; then \
			echo "📁 前端构建产物不存在，开始构建..."; \
			FRONTEND_CHANGED=1; \
		else \
			echo "✅ 前端文件无变化，跳过构建"; \
		fi; \
		if [ $$FRONTEND_CHANGED -eq 1 ]; then \
			echo "🚀 正在构建前端项目..."; \
			cd admin && ./build > /dev/null 2>&1 && echo "✅ 前端构建完成" || (echo "❌ 前端构建失败" && exit 1); \
			cd ..; \
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

# 忽略不存在的目标（用于处理额外参数）
.PHONY: $(filter-out $(firstword $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(filter-out $(firstword $(MAKECMDGOALS)),$(MAKECMDGOALS)):
	@: