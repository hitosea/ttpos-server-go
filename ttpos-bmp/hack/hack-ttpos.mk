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
	@echo "通过mysql客户端容器创建 erp, takeout 数据库"
	@./hack/init_db.sh

# 重新构建所有应用并启动
.PHONY: up
up:
	@# 指定项目名 -p ttpos-bmp
	@set -o allexport; \
	. ../.env && docker compose  -p ttpos-bmp -f ./docker-compose.yml up -d --build;\
	set +o allexport;


# 仅启动中间件
.PHONY: mid
mid:
	@set -o allexport; \
	. ../.env && docker compose -p ttpos-bmp-mid -f ./docker-compose.mid.yml up -d ;\
	set +o allexport;

# 启动中间件及中台应用服务
.PHONY: run
run:
	@# 指定项目名 -p ttpos-bmp
	@set -o allexport; \
	. ../.env && docker compose -f ./docker-compose.yml -f ./docker-compose.mid.yml up -d ;\
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

# 迁移升级所有应用的数据库
.PHONY: migrate
migrate:
	@make create_db
	@cd app/ttpos-takeout  && make db_up.docker
	#@cd app/ttpos-manager  && make db_up.docker
	#@cd app/ttpos-shop  && make db_up.docker
	@cd app/ttpos-erp  && make db_up.docker