
# Install/Update to the latest CLI tool.
.PHONY: cli
cli:
	@set -e; \
	wget -O gf \
		https://github.com/gogf/gf/releases/latest/download/gf_$(shell go env GOOS)_$(shell go env GOARCH) && \
	chmod +x gf && \
	./gf install -y && \
	rm ./gf


# Check and install CLI tool.
.PHONY: cli.install
cli.install:
	@set -e; \
	gf -v > /dev/null 2>&1 || if [[ "$?" -ne "0" ]]; then \
  		echo "GoFame CLI is not installed, start proceeding auto installation..."; \
		make cli; \
	fi;



# 启动 Docker 容器（使用项目名 ttpos-bmp 并加载 ../.env）
.PHONY: docker
docker:
	@# 指定项目名 -p ttpos-bmp
	@set -o allexport; \
	source ../.env && docker compose -p ttpos-bmp -f ./docker-compose.yml up -d ;\
	set +o allexport


# 构建并运行 ttpos-manager 服务
.PHONY: run.manager
run.manager:
	@cd app/ttpos-manager && gf run main.go

# 构建并运行 ttpos-shop 服务
.PHONY: run.shop
run.shop:
	@cd app/ttpos-shop && gf run main.go:
