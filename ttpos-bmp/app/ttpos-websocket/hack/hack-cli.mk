# GoFrame CLI tool installation
.PHONY: cli.install
cli.install:
	@if ! command -v gf >/dev/null 2>&1; then \
		echo "Installing gf CLI tool..."; \
		go install github.com/gogf/gf/cmd/gf/v2@latest; \
	fi

.PHONY: migrate.install
migrate.install:
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Installing migrate tool..."; \
		go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi

