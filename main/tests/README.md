# Integration Tests

Integration tests for the TTPOS server. Tests run inside Docker containers against real MySQL, Redis, and a WireMock gRPC stub.

## Quick Start

```bash
# Run tests locally (from repo root)
make test-local

# Clean up containers and coverage
make test-clean
```

## Architecture

```
docker-compose.yml
  |
  +-- mysql-test          MySQL 8.0 (saas.sql init, healthcheck on table)
  +-- redis-test          Redis 6.0
  +-- wiremock            WireMock 3.9.1 (mocks gRPC BMP services)
  +-- ttpos-server-go     Go server built with -cover (Dockerfile.coverage)
  +-- test-runner         Go test binary (tests/Dockerfile)
```

**Startup order:** mysql-test (healthy) -> redis-test + wiremock (started) -> ttpos-server-go (healthy) -> test-runner

## Coverage Pipeline

1. Server binary built with `go build -cover` writes raw coverage to `/coverage` volume
2. After tests, `go tool covdata textfmt` converts binary data to `coverage/total.out`
3. `fix-coverage-paths.sh` transforms Go module paths to filesystem paths for SonarQube
4. In CI, unit + integration coverage are merged with `gocovmerge`

### Why fix-coverage-paths.sh?

Go coverage records paths as `ttpos-server-go/app/...` (module name from go.mod), but SonarQube expects `main/app/...` (filesystem path). The script reads the module name from `go.mod` dynamically to avoid hardcoding.

## Environment Variables

Default values are in [`test.env`](env.test). Override by exporting before running:

```bash
export DB_ROOT_PASSWORD=mypassword
make test-local
```

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_ROOT_PASSWORD` | `testroot` | MySQL root password |
| `DB_DATABASE` | `saas` | SaaS main database name |
| `DB_USERNAME` | `test` | MySQL test user |
| `DB_PASSWORD` | `test` | MySQL test password |
| `JWT_SECRET` | `test-secret-key-...` | JWT signing key |

## Adding New Tests

1. Create a new package under `main/tests/` (e.g., `main/tests/member/`)
2. Use build tag `//go:build integration` at the top of test files
3. Use fixtures from `main/tests/fixture/`:
   - `fixture.DefaultDBConfig()` — database connection
   - `fixture.NewTestTenantFull(t, uuid)` — creates a tenant DB with full schema, auto-cleaned up
   - `fixture.NewHTTPClient(t, serviceURL)` — HTTP client with assertion helpers
   - `fixture.GenerateToken(staffUUID, companyUUID, secret)` — JWT for auth
4. Test files must use `_test.go` suffix and `-tags=integration` build tag

Example:

```go
//go:build integration

package member_test

import (
    "testing"
    "main/tests/fixture"
)

func TestMember_Create(t *testing.T) {
    cfg := fixture.DefaultDBConfig()
    db := fixture.NewTestTenantFull(t, "9999000000000001")
    client := fixture.NewHTTPClient(t, fixture.ServiceURL())
    // ... test logic
}
```

## Files

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Service definitions and healthchecks |
| `Dockerfile` | Test runner image |
| `test.env` | Default environment variables |
| `fix-coverage-paths.sh` | Coverage path transform for SonarQube |
| `sonar-project.properties` | SonarQube scanner config |
| `fixture/` | Shared test helpers (DB, HTTP, auth, seeding) |
| `order/` | Order integration tests |
