# BMP Integration Test Runner
# ============================
# Runs integration tests against bmp-erp, bmp-takeout, and bmp-message services.
# Build context is the repository root.

FROM golang:1.23-alpine

RUN apk add --no-cache bind-tools

WORKDIR /app

# Copy go.mod/go.sum first for Docker layer caching
COPY ttpos-bmp/go.mod ttpos-bmp/go.sum ./ttpos-bmp/
COPY ttpos-api/go.mod ttpos-api/go.sum ./ttpos-api/

# Create a minimal tests module
COPY ttpos-bmp/tests/go.mod ttpos-bmp/tests/go.sum* ./ttpos-bmp/tests/
RUN cd ttpos-bmp/tests && go mod download 2>/dev/null || true

# Copy all source
COPY . ./

WORKDIR /app/ttpos-bmp/tests

CMD ["sh", "-c", "go test -tags=integration -v -count=1 -timeout=10m ./..."]
