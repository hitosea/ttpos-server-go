
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


# Install protobuf tools with fixed versions.
.PHONY: protobuf.install
protobuf.install:
	@echo "Installing protobuf tools with fixed versions..."
	@bash hack/install-protobuf-tools.sh


# Check protobuf tools versions.
.PHONY: protobuf.version
protobuf.version:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Protobuf Tools Version Check"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo -n "protoc:             "
	@protoc --version 2>/dev/null || echo "NOT INSTALLED"
	@echo -n "protoc-gen-go:      "
	@protoc-gen-go --version 2>/dev/null || echo "NOT INSTALLED"
	@echo -n "protoc-gen-go-grpc: "
	@protoc-gen-go-grpc --version 2>/dev/null || echo "NOT INSTALLED"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
