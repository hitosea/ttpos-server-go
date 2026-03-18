## Why

The TTPOS server has no integration test infrastructure. The codebase only has scattered unit tests using SQLite in-memory DBs that miss MySQL-specific issues. Additionally, 7 docker-compose files across the repo create confusion and config drift. We need an integration test framework that:
- Validates service correctness against **production images** from the registry
- Enables performance bottleneck discovery via CPU profiling
- Supports **parallel test runs** with complete isolation
- Consolidates Docker infrastructure with profile-based compose files

## What Changes

- **New unified Docker Compose**: Two compose files replace 7 scattered files:
  - `docker-compose.yml` at repo root with `dev`/`test`/`prod` profiles for Main services
  - `ttpos-bmp/docker-compose.yml` with `mid`/`dev`/`prod` profiles for BMP services
  - **Test profile**: Adds MySQL + WireMock + **test-runner** containers
  - Service runs as **production image** pulled from registry (no builds in CI)
  - Tests run inside Docker container (network-isolated, not on host)
  - **New test framework library** (`main/testlib/`):
  - WireMock Admin API client for stub registration/verification
  - Test fixture helpers using `testing.TB` interface for benchmark reuse
  - **New scenario-based test directory** (`main/tests/`):
  - Integration tests organized by business domain (order, desk, member, auth)
  - Tests run as HTTP clients against running service container
  - Each test run is fully isolated via Docker Compose project names
  - **Makefile targets**: `test-integration`, `test-clean` for running and cleaning up tests

## Capabilities

### New Capabilities
- `docker-compose-profiles`: Unified Docker Compose infrastructure with profile-based environment selection, isolated test runs via project names
- `test-framework-core`: Go test framework library providing WireMock client and fixture helpers
- `integration-test-scenarios`: Scenario-based integration test suite with HTTP client tests running in isolated Docker containers

### Modified Capabilities

## Impact

- **Docker infrastructure**: All 7 existing compose files replaced by 2 profile-based files. Commands change to use `--profile` flag.
- **Production code**: **No modifications needed** — service uses standard DBManager and config, runs as production image
- **Dependencies**:
  - `test-runner` container needs Go toolchain (small image)
  - WireMock gRPC extension JAR for proto mocking
- **CI pipeline**:
  - Pulls production image from registry
  - Runs `docker compose -p test-$BUILD_ID --profile test up --abort-on-container-exit`
  - Test results from test-runner container exit code
- **go.mod**: No changes needed
- **Network isolation**: Multiple parallel test runs supported via Docker Compose project names (e.g., `test-123`, `test-456`)
