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
# Keep the default package-level parallelism. By default we keep tenant DBs
# until compose teardown, because some async integration paths may still touch
# them briefly after the test returns.
CMD ["sh", "-c", \
  "export TEST_DROP_TENANT_DB=${TEST_DROP_TENANT_DB:-0}; \
   go test -p ${GO_TEST_P:-4} -tags=integration -v -count=1 -timeout=10m ./..."]
