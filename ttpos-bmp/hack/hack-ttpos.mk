# 加载环境变量并替换数据库连接配置
.PHONY: conf
conf:
	@sh hack/init_conf.sh

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


# 启动 Docker 容器（使用项目名 ttpos-bmp 并加载 ../.env）
.PHONY: docker
docker:
	@# 指定项目名 -p ttpos-bmp
	@set -o allexport; \
	source ../.env && docker compose -p ttpos-bmp -f ./docker-compose.yml up -d ;\
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