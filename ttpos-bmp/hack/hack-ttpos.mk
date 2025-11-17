# 加载环境变量并替换数据库连接配置
.PHONY: conf
conf:
	@/bin/bash hack/init_conf.sh

# Check and install envsubst.
.PHONY: envsubst.install
envsubst.install:
	@set -e; \
	envsubst -V > /dev/null 2>&1 || if [[ "$?" -ne "0" ]]; then \
		echo "envsubst is not installed, start proceeding auto installation..."; \
		brew install gettext; \
	fi;

.PHONY: migrate.install
migrate.install:
	@set -e; \
	migrate -version > /dev/null 2>&1 || if [[ "$?" -ne "0" ]]; then \
		echo "migrate is not installed, start proceeding auto installation..."; \
		go install -tags mysql github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi;


.PHONY: create_db
create_db:
	@echo "通过mysql客户端容器创建 erp, takeout, message 数据库"
	@./hack/init_db.sh

# 重新构建所有应用并启动
.PHONY: up
up:
	@# 指定项目名 -p ttpos-bmp
	@set -o allexport; \
	. ../.env && docker compose  -p ttpos-bmp -f ./docker-compose-dev.yml up -d --build;\
	set +o allexport;


# 仅启动中间件
.PHONY: mid
mid:
	@set -o allexport; \
	. ../.env && docker compose -p ttpos-bmp-mid -f ./docker-compose.mid.yml up -d ;\
	set +o allexport;
	@make update-topic

# 启动中间件及中台应用服务
.PHONY: run
run:
	@# 指定项目名 -p ttpos-bmp
	@set -o allexport; \
	. ../.env && docker compose -f ./docker-compose-dev.yml -f ./docker-compose.mid.yml up -d ;\
	set +o allexport;

# 构建并运行 ttpos-manager 服务
.PHONY: run.manager
run.manager:
	@cd app/ttpos-manager && gf run main.go ;

# 构建并运行 ttpos-shop 服务
.PHONY: run.shop
run.shop:
	@cd app/ttpos-shop && gf run main.go ;

# 构建并运行 ttpos-erp 服务
.PHONY: run.erp
run.erp:
	@cd app/ttpos-erp && gf run main.go ;

 # 构建并运行 ttpos-takeout 服务
.PHONY: run.takeout
run.takeout:
	@cd app/ttpos-takeout && gf run main.go

 # 构建并运行 ttpos-message 服务
.PHONY: run.message
run.message:
	@cd app/ttpos-message && gf run main.go

# 迁移升级所有应用的数据库
.PHONY: migrate
migrate:
	@make create_db
	@cd app/ttpos-takeout  && make db_up.docker
	#@cd app/ttpos-manager  && make db_up.docker
	#@cd app/ttpos-shop  && make db_up.docker
	@cd app/ttpos-erp  && make db_up.docker
	@cd app/ttpos-message  && make db_up.docker
	@make update-topic


# 替换本地IP， 开发用
.PHONY: update-ip
update-ip:
	@./hack/update_ip.sh

# 执行ERP数据迁移，初始化自定义字段、客户和支付方式
# 使用方法: make erp.migrate SITE_CODE=1 DIR_BASE=./manifest/erp-migrate/v2.5
# 参数说明:
#   SITE_CODE: ERP站点代码，默认为 1
#   DIR_BASE:  迁移文件目录路径，默认为 ./manifest/erp-migrate/v2.5
.PHONY: erp.migrate
erp.migrate:
	@if [ -z "$(SITE_CODE)" ]; then \
		echo "❌ 错误: 请提供 SITE_CODE 参数"; \
		echo "📖 使用方法: make erp.migrate SITE_CODE=1 DIR_BASE=./manifest/erp-migrate/v2.5"; \
		echo "📋 示例:"; \
		echo "   make erp.migrate SITE_CODE=1 DIR_BASE=./manifest/erp-migrate/v2.5"; \
		echo "   make erp.migrate SITE_CODE=2 DIR_BASE=./manifest/erp-migrate/payment"; \
		exit 1; \
	fi
	@if [ -z "$(DIR_BASE)" ]; then \
		echo "❌ 错误: 请提供 DIR_BASE 参数"; \
		echo "📖 使用方法: make erp.migrate SITE_CODE=1 DIR_BASE=./manifest/erp-migrate/v2.5"; \
		echo "📋 示例:"; \
		echo "   make erp.migrate SITE_CODE=1 DIR_BASE=./manifest/erp-migrate/v2.5"; \
		echo "   make erp.migrate SITE_CODE=2 DIR_BASE=./manifest/erp-migrate/payment"; \
		exit 1; \
	fi
	@echo "🚀 开始执行ERP数据迁移..."
	@echo "📍 站点代码: $(SITE_CODE)"
	@echo "📁 迁移目录: $(DIR_BASE)"
	@cd app/ttpos-erp && gf run main.go --args "migrate --siteCode $(SITE_CODE) --dirBase $(DIR_BASE)"
	@echo "✅ ERP数据迁移执行完成!"


# 执行所有版本目录的ERP数据迁移
# 使用方法: make erp.migrate-all SITE_CODE=1 BASE_DIR=./manifest/erp-migrate
# 参数说明:
#   SITE_CODE: ERP站点代码，默认为 1
#   BASE_DIR:  迁移根目录路径，默认为 ./manifest/erp-migrate
.PHONY: erp.migrate-all
erp.migrate-all:
	@if [ -z "$(SITE_CODE)" ]; then \
		echo "❌ 错误: 请提供 SITE_CODE 参数"; \
		echo "📖 使用方法: make erp.migrate-all SITE_CODE=1 BASE_DIR=./manifest/erp-migrate"; \
		echo "📋 示例:"; \
		echo "   make erp.migrate-all SITE_CODE=1 BASE_DIR=./manifest/erp-migrate"; \
		exit 1; \
	fi
	@if [ -z "$(BASE_DIR)" ]; then \
		echo "❌ 错误: 请提供 BASE_DIR 参数"; \
		echo "📖 使用方法: make erp.migrate-all SITE_CODE=1 BASE_DIR=./manifest/erp-migrate"; \
		echo "📋 示例:"; \
		echo "   make erp.migrate-all SITE_CODE=1 BASE_DIR=./manifest/erp-migrate"; \
		exit 1; \
	fi
	@echo "🚀 开始执行ERP全量数据迁移..."
	@echo "📍 站点代码: $(SITE_CODE)"
	@echo "📁 迁移根目录: $(BASE_DIR)"
	@cd app/ttpos-erp && gf run main.go --args "migrate-all --siteCode $(SITE_CODE) --dirBase $(BASE_DIR)"
	@echo "✅ ERP全量数据迁移执行完成!"


# 更新所有topic
.PHONY: update-topic
update-topic:
	@./hack/update_topic.sh
