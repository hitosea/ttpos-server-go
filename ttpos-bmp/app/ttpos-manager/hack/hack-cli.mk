
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
