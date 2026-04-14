.DEFAULT_GOAL := build

# Update GoFrame and its CLI to latest stable version.
.PHONY: up
up: cli.install
	@gf up -a

# Build binary using configuration from hack/config.yaml.
.PHONY: build
build: cli.install
	@gf build -ew

# Parse api and generate controller/sdk.
.PHONY: ctrl
ctrl: cli.install
	@gf gen ctrl

# Generate Go files for DAO/DO/Entity.
.PHONY: dao
dao: cli.install
	@gf gen dao

# Parse current project go files and generate enums go file.
.PHONY: enums
enums: cli.install
	@gf gen enums

# Generate Go files for Service.
.PHONY: service
service: cli.install
	@gf gen service


# Build docker image.
.PHONY: image
image: cli.install
	$(eval _TAG  = $(shell git rev-parse --short HEAD))
ifneq (, $(shell git status --porcelain 2>/dev/null))
	$(eval _TAG  = $(_TAG).dirty)
endif
	$(eval _TAG  = $(if ${TAG},  ${TAG}, $(_TAG)))
	$(eval _PUSH = $(if ${PUSH}, ${PUSH}, ))
	@gf docker ${_PUSH} -tn $(DOCKER_NAME):${_TAG};


# Build docker image and automatically push to docker repo.
.PHONY: image.push
image.push: cli.install
	@make image PUSH=-p;


# Deploy image and yaml to current kubectl environment.
.PHONY: deploy
deploy: cli.install
	$(eval _TAG = $(if ${TAG},  ${TAG}, develop))

	@set -e; \
	mkdir -p $(ROOT_DIR)/temp/kustomize;\
	cd $(ROOT_DIR)/manifest/deploy/kustomize/overlays/${_ENV};\
	kustomize build > $(ROOT_DIR)/temp/kustomize.yaml;\
	kubectl   apply -f $(ROOT_DIR)/temp/kustomize.yaml; \
	if [ $(DEPLOY_NAME) != "" ]; then \
		kubectl patch -n $(NAMESPACE) deployment/$(DEPLOY_NAME) -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"$(shell date +%s)\"}}}}}"; \
	fi;


# Parsing protobuf files and generating go files.
.PHONY: pb
pb: cli.install
	@gf gen pb

# Generate protobuf files for database tables.
.PHONY: pbentity
pbentity: cli.install
	@gf gen pbentity


.PHONY: db_add
db_add: migrate.install
	@migrate create -ext sql -dir ./manifest/sql -tz Asia/Shanghai $(NAME)

.PHONY: db_up
db_up: migrate.install
	$(eval _LINK = $(shell yq '.migrate-db-link' hack/config.yaml))
	@migrate -verbose -path ./manifest/sql -database "$(_LINK)" up

.PHONY: db_up.docker
db_up.docker:
	@# 兼容两种部署：DB_DOCKER_NETWORK 非空则加入 docker 网络用内部端口 DB_PORT；否则用宿主端口 DB_PORT_OPEN
	@ENV_FILE=$(ROOT_DIR)/../../.env ; [ -f "$$ENV_FILE" ] || ENV_FILE=$(ROOT_DIR)/../../../.env ; \
	set -a ; . "$$ENV_FILE" ; set +a ; \
	DB_DSN=$$(sed -n 's/migrate-db-link-open:[[:space:]]*\(.*\)/\1/p' $(ROOT_DIR)/hack/config.yaml | sed 's/"//g') ; \
	if [ -z "$$DB_DSN" ]; then echo "未在 config.yaml 中找到有效的 link 配置" >&2; exit 1; fi ; \
	if [ -n "$$DB_DOCKER_NETWORK" ]; then \
		NETWORK_OPT="--network $$DB_DOCKER_NETWORK" ; \
		DB_DSN=$$(echo "$$DB_DSN" | sed "s/:$$DB_PORT_OPEN)/:$$DB_PORT)/g") ; \
	else \
		NETWORK_OPT="" ; \
	fi ; \
	echo "DSN: $$DB_DSN" ; \
	docker run --rm $$NETWORK_OPT -v $(ROOT_DIR)/manifest/sql:/migrations migrate/migrate -path /migrations -database "$$DB_DSN" up

.PHONY: db_down
db_down: migrate.install
	$(eval _LINK = $(shell yq '.migrate-db-link' hack/config.yaml))
	@migrate -verbose -path ./manifest/sql -database "$(_LINK)" down

.PHONY: db_force
db_force: migrate.install
	$(eval _LINK = $(shell yq '.migrate-db-link' hack/config.yaml))
	@migrate -path ./manifest/sql -database "$(_LINK)" force $(VERSION)

.PHONY: db_version
db_version: migrate.install
	$(eval _LINK = $(shell yq '.migrate-db-link' hack/config.yaml))
	@migrate -path ./manifest/sql -database "$(_LINK)" version

