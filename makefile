install:
	cd admin && ./cmd init

run:
	cd admin && ./cmd mysql open
	cd main && go mod tidy
	cd main && go run ./cmd/server/main.go ./cmd/server/swagger_enabled.go

dev:
	cd admin && ./cmd mysql open
	if [ ! -f "${HOME}/go/bin/fresh" ]; then \
		go install github.com/pilu/fresh@latest; \
	fi
	cd main && ${HOME}/go/bin/fresh -c ./fresh.conf