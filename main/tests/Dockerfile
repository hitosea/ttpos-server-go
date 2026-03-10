# TTPOS Integration Test Runner
# ==============================
# This container runs integration tests against the ttpos-server-go service.
#
# Build context should be repository root to access all modules.

FROM golang:1.23-alpine

# Install dnsutils for DNS resolution check
RUN apk add --no-cache bind-tools

WORKDIR /app

# Copy go.mod/go.sum first for Docker layer caching
COPY main/go.mod main/go.sum ./main/
COPY ttpos-bmp/go.mod ttpos-bmp/go.sum ./ttpos-bmp/
COPY ttpos-api/go.mod ttpos-api/go.sum ./ttpos-api/
RUN cd main && go mod download

# Copy all source files (build context is repository root)
COPY . ./

WORKDIR /app/main/tests

# Blackbox API tests only — unit tests (pkg/lock, pkg/cache, etc.) belong in
# their own CI stage, not the integration test phase.
# All packages run concurrently (default -p). Each test creates an isolated
# tenant DB with a unique company_uuid, so there are no cross-package DB
# collisions. WireMock default stubs are registered once via fixture.SetupWireMock
# (sync.Once) and never reset.
CMD ["sh", "-c", \
  "go test -tags=integration -v -count=1 -timeout=10m ./..."]
