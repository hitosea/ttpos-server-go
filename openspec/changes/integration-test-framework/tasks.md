## 1. Docker Compose Consolidation

- [x] 1.1 Create `docker-compose.yml` at repo root with `dev`, `test`, `prod` profiles. Port all services from existing files. **Test profile includes: mysql, wiremock, ttpos-server-go (production image), test-runner. No port mapping to host — internal Docker DNS only.**
- [ ] 1.2 Create `ttpos-bmp/docker-compose.yml` with `mid`, `dev`, `prod` profiles — port all services from existing files.
- [x] 1.3 Create `.env.test` with test-specific env vars (JWT secret, DB credentials, WireMock config).
- [ ] 1.4 Verify all profiles work: `--profile dev`, `--profile test`, `--profile prod` for Main; `--profile mid`, `--profile dev`, `--profile prod` for BMP.
- [x] 1.5 Add deprecation comments to old compose files pointing to new ones.

## 2. Test Runner Container

- [x] 2.1 Create `main/tests/Dockerfile` — small Go image (`golang:1.24-alpine`) with test source mounted or copied.
- [x] 2.2 Add `test-runner` service to docker-compose.yml test profile — builds from `main/tests/Dockerfile`, depends on ttpos-server-go, mysql, wiremock, connects via Docker DNS.
- [x] 2.3 Configure `test-runner` environment: `SERVICE_URL=http://ttpos-server-go:8080`, `DB_HOST=mysql`, `WIREMOCK_URL=http://wiremock:8080`, `WIREMOCK_GRPC=wiremock:8443`.
- [x] 2.4 Set `test-runner` command to `go test -tags=integration -v -count=1 -timeout=10m ./tests/...` and use `--abort-on-container-exit` pattern.
- [x] 2.5 Configure health checks and dependencies:
  ```yaml
  mysql:
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 5s
      retries: 10

  test-runner:
    depends_on:
      mysql:
        condition: service_healthy
      wiremock:
        condition: service_started
      ttpos-server-go:
        condition: service_started
  ```

## 3. Service Container Configuration

- [x] 3.1 Configure `ttpos-server-go` in test profile to use production image:
  ```yaml
  ttpos-server-go:
    image: hub.hitosea.com/ttpos-server/ttpos-server-go:${IMAGE_TAG:-latest}
    profiles: [test]
    stop_grace_period: 10s
    stop_signal: SIGTERM
  ```
- [x] 3.2 Set service environment variables to point to WireMock: `GRPC_ERP_ADDR=wiremock:8443`, `GRPC_TAKEOUT_ADDR=wiremock:8443`, etc.
- [x] 3.3 For local development, add alternative service definition that builds from source:
  ```yaml
  ttpos-server-go-local:
    build: { context: main, dockerfile: Dockerfile }
    profiles: [test-local]
  ```

## 4. WireMock Setup

- [x] 4.1 Add WireMock service to docker-compose.yml test profile with gRPC extension:
  ```yaml
  wiremock:
    image: wiremock/wiremock:3.x
    profiles: [test, test-coverage]
    command: --port 8080 --https-port 8443 --extensions org.wiremock.grpc.GrpcExtension
    volumes:
      - ./main/testlib/wiremock/extensions:/var/wiremock/extensions
      - ./main/testlib/wiremock/descriptors:/var/wiremock/descriptors
  ```
- [x] 4.2 Create `main/testlib/wiremock/descriptors/` directory for proto descriptor sets.
- [x] 4.3 Create `main/testlib/Makefile` with `proto-desc` target to compile all protos into `all_services.desc`.
- [ ] 4.4 Download WireMock gRPC extension JAR into `main/testlib/wiremock/extensions/`.
- [x] 4.5 Create `main/testlib/wiremock/defaults.go` — registers catch-all no-op stubs for all gRPC services (ERP, Takeout, Message, WebSocket) to handle GORM model hooks that make gRPC calls.

## 5. WireMock Client Library (`main/testlib/wiremock/`)

- [x] 5.1 Create `stub.go` — type definitions: `StubMapping`, `RequestPattern`, `ResponseDef`, `Matcher`.
- [x] 5.2 Create `client.go` — WireMock Admin API client: `NewClient(baseURL)`, `RegisterStub()`, `DeleteStub()`, `Verify()`.
- [x] 5.3 Create `http_builder.go` — fluent HTTP stub builder with `WithTestCaseID()`.
- [x] 5.4 Create `grpc_builder.go` — fluent gRPC stub builder for WireMock gRPC extension.

## 6. Test Fixtures (`main/testlib/fixture/`)

- [x] 6.1 Create `db.go` — `NewTestTenant(tb)` creates `shop{uuid}` database via DB_HOST env var, registers cleanup. All functions accept `testing.TB`.
- [x] 6.2 Create `auth.go` — `GenerateToken(AuthConfig)` creates JWT using JWT_SECRET env var.
- [x] 6.3 Create `seed.go` — `SeedStaff(tb, db, ...overrides)`, `SeedDesk()`, `SeedDevice()` with override pattern.
- [x] 6.4 Create `http.go` — `NewHTTPClient()` creates client configured for SERVICE_URL env var.

## 7. Example Integration Test

- [x] 7.1 Create `main/tests/order/happy_path_test.go` — HTTP client test that:
  - Creates tenant via `fixture.NewTestTenant()`
  - Seeds data via `fixture.SeedStaff()`, `fixture.SeedDevice()`
  - Registers WireMock stub via `wiremock.NewGrpcStub()`
  - Makes HTTP POST to `SERVICE_URL/api/v1/cashier/order/create`
  - Asserts response status and body
  - Verifies WireMock was called
  - Cleans up stub and tenant database

## 8. Verification

- [ ] 8.1 Run `docker compose -p test-manual --profile test up -d` — verify all containers start.
- [ ] 8.2 Run `docker compose -p test-manual --profile test up --abort-on-container-exit` — verify test-runner executes and passes.
- [ ] 8.3 Run two parallel test runs with different project names — verify isolation.
- [ ] 8.4 Verify CI can pull production image and run tests against it.

## 9. Code Coverage Support

- [x] 9.1 Create `main/Dockerfile.coverage` — builds service binary with `go build -cover` flag.
- [x] 9.2 Add `test-coverage` profile to docker-compose.yml:
  ```yaml
  ttpos-server-go-coverage:
    build:
      context: main
      dockerfile: Dockerfile.coverage
    profiles: [test-coverage]
    stop_grace_period: 10s
    stop_signal: SIGTERM
    volumes:
      - ./coverage:/coverage
    environment:
      - GOCOVERDIR=/coverage
  ```
- [x] 9.3 Ensure coverage output directory `coverage/` is writable by container.

## 10. SonarQube Configuration

- [x] 10.1 Create `sonar-project.properties` at repo root (minimal - Jenkins passes most params):
  ```properties
  sonar.projectKey=ttpos-server-go
  sonar.sources=main
  sonar.tests=main
  sonar.test.inclusions=**/*_test.go
  sonar.coverage.exclusions=**/dao/**,**/entity/**,**/pb/**,**/testlib/**
  sonar.exclusions=ttpos-bmp/**/dao/**,ttpos-bmp/**/model/entity/**,ttpos-bmp/**/model/do/**,**/pb/**
  ```

## 11. Makefile API (invoked by Jenkins)

- [x] 11.1 Create root `Makefile` with test execution targets only:
  ```makefile
  BUILD_ID ?= $(shell date +%s)

  .PHONY: test-integration test-clean

  test-integration:
  	docker compose -p test-$(BUILD_ID) --profile test up --abort-on-container-exit && \
  	docker compose -p test-$(BUILD_ID) down -v

  test-clean:
  	docker compose -p test-* down -v 2>/dev/null || true
  ```

  Jenkins pipeline handles: unit tests, coverage merge, SonarQube upload.
