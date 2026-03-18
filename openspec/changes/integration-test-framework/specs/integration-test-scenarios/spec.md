## ADDED Requirements

### Requirement: Scenario-based test directory structure
The system SHALL organize integration tests in `main/tests/` with subdirectories per business domain. Each subdirectory SHALL be its own Go package.

#### Scenario: Test directory layout
- **WHEN** a developer looks at the test directory
- **THEN** they find subdirectories organized by business domain (e.g., `order/`, `desk/`, `member/`, `auth/`) each containing scenario test files

#### Scenario: Tests run in test-runner container
- **WHEN** Docker Compose starts the test profile
- **THEN** a `test-runner` container runs `go test -tags=integration ./tests/...` and exits with the test result code

### Requirement: Build tag separation
All integration test files SHALL use the `//go:build integration` build tag. Unit tests SHALL NOT have this tag. The two test suites SHALL never compile together.

#### Scenario: Unit tests exclude integration tests
- **WHEN** `go test ./...` is run without `-tags=integration`
- **THEN** no files with `//go:build integration` are compiled or executed

#### Scenario: Integration tests run with tag
- **WHEN** `go test -tags=integration ./tests/...` is run inside test-runner container
- **THEN** all integration test files are compiled and executed

### Requirement: Parallel execution with UUID isolation
Every integration test function SHALL call `t.Parallel()` and use a unique tenant (via `fixture.NewTestTenant`) to ensure zero shared state between tests.

#### Scenario: Parallel tests do not interfere
- **WHEN** 20 tests run in parallel, each creating orders in their own tenant database
- **THEN** no test fails due to data from another test, and all WireMock stubs are isolated by test case ID

### Requirement: HTTP client tests against running service
Tests SHALL make real HTTP requests to the running `ttpos-server-go` container. Tests SHALL NOT use in-process `engine.ServeHTTP()` calls.

#### Scenario: Test makes HTTP request
- **WHEN** a test calls `http.Post(SERVICE_URL+"/api/v1/order/create", body)`
- **THEN** the request goes to the running service container via Docker DNS (`ttpos-server-go:8080`)

#### Scenario: Test receives real HTTP response
- **WHEN** the service processes the request
- **THEN** the test receives the actual HTTP response with status code, headers, and body

### Requirement: Docker Compose project isolation
Each test run SHALL use a unique Docker Compose project name (`-p test-$BUILD_ID`) to enable parallel test runs with complete isolation.

#### Scenario: Parallel CI runs are isolated
- **WHEN** CI runs `docker compose -p test-123 --profile test up` and `docker compose -p test-456 --profile test up` simultaneously
- **THEN** both runs have separate networks, containers, and databases with zero conflicts

### Requirement: Service uses production image in CI
The service container in CI SHALL use the production image from the registry (`hub.hitosea.com/.../ttpos-server-go:${IMAGE_VERSION}`), not a local build.

#### Scenario: CI pulls and tests production image
- **WHEN** CI runs the test pipeline
- **THEN** it pulls `hub.hitosea.com/ttpos-server/ttpos-server-go:${IMAGE_VERSION}` and runs tests against that exact image

#### Scenario: Local development builds from source
- **WHEN** a developer runs tests locally
- **THEN** the service container is built from local source code (`build: { context: main }`)

### Requirement: Makefile targets for test workflow
The system SHALL provide Makefile targets that integrate Docker Compose lifecycle with test execution.

#### Scenario: Start test infrastructure
- **WHEN** developer runs `make test-up` in the `main/` directory
- **THEN** Docker Compose starts the `test` profile services with a unique project name

#### Scenario: Run integration tests
- **WHEN** developer runs `make test-integration`
- **THEN** Docker Compose runs the test-runner container which executes `go test -tags=integration`

#### Scenario: Full pipeline
- **WHEN** developer runs `make test` or `make test-all`
- **THEN** the system runs `test-up`, then `test-integration`, then `test-down` in sequence

### Requirement: Inline struct test data
Integration tests SHALL use inline Go structs for test data. External data files (YAML, JSON) SHALL NOT be used unless the data exceeds 1000 lines.

#### Scenario: Test uses inline data
- **WHEN** a test needs to create an order with products
- **THEN** the test defines the request body and seed data as Go struct literals or string literals within the test function

### Requirement: Example order happy path test
The system SHALL include at least one working example integration test in `main/tests/order/happy_path_test.go` that demonstrates the full test pattern: tenant setup, data seeding, WireMock stub registration, HTTP request to running service, response assertion, and mock verification.

#### Scenario: Example test passes
- **WHEN** Docker Compose runs the test profile
- **THEN** the test-runner container creates a tenant, seeds data, registers a WireMock stub, sends an HTTP request to the service container, and verifies the response and mock call count

### Requirement: CI pipeline integration
The integration test suite SHALL be runnable in CI with Docker Compose project isolation.

#### Scenario: CI runs isolated tests
- **WHEN** a CI pipeline executes `docker compose -p test-$BUILD_ID --profile test up --abort-on-container-exit`
- **THEN** integration tests run in isolation and the exit code reflects test results

### Requirement: Performance profiling support
The test framework SHALL support CPU profiling on any integration test for bottleneck discovery.

#### Scenario: CPU profile a specific test
- **WHEN** test-runner runs `go test -tags=integration -run TestOrderHappyPath -cpuprofile=cpu.out ./tests/order/...`
- **THEN** a `cpu.out` file is generated that can be analyzed with `go tool pprof`
