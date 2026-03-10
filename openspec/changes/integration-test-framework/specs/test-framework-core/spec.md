## ADDED Requirements

### Requirement: WireMock Admin API client
The system SHALL provide a Go package `testlib/wiremock/` with a client that wraps the WireMock Admin API for stub registration, deletion, and verification.

#### Scenario: Register an HTTP stub
- **WHEN** a test calls `client.RegisterStub(stub)` with an HTTP stub mapping
- **THEN** the stub is registered in WireMock and a stub UUID is returned for later cleanup

#### Scenario: Register a gRPC stub
- **WHEN** a test calls `client.RegisterStub(stub)` with a gRPC stub mapping built via `NewGrpcStub(service, method)`
- **THEN** the gRPC stub is registered in WireMock using the gRPC extension format

#### Scenario: Delete a stub by UUID
- **WHEN** a test calls `client.DeleteStub(stubUUID)`
- **THEN** the stub is removed from WireMock

#### Scenario: Verify stub call count
- **WHEN** a test calls `client.Verify(pattern, expectedCount)`
- **THEN** the client returns nil if the request pattern was matched exactly `expectedCount` times, or an error with details otherwise

### Requirement: HTTP stub builder with test isolation
The system SHALL provide a fluent builder `NewHTTPStub()` that enforces test-case-level isolation via `WithTestCaseID(id)`.

#### Scenario: Build HTTP stub with test case ID
- **WHEN** a test builds a stub with `NewHTTPStub().ForMethod("POST").ForPath("/api/payment").WithTestCaseID(t.Name()).WillReturnJSON(200, body).Build()`
- **THEN** the built StubMapping includes a header matcher for `X-Test-ID` equal to `t.Name()` and metadata tagging with the test case ID

### Requirement: gRPC stub builder
The system SHALL provide a fluent builder `NewGrpcStub(serviceName, methodName)` that produces WireMock gRPC extension-compatible stub mappings.

#### Scenario: Build gRPC stub with JSON response
- **WHEN** a test builds a stub with `NewGrpcStub("erp.SellingService", "SavePosInvoice").WithTestCaseID(t.Name()).WillReturn(responseMap).Build()`
- **THEN** the built StubMapping uses the WireMock gRPC extension format with the correct service/method path and JSON-encoded response body

### Requirement: Test fixture configuration via environment
The system SHALL provide fixture helpers that read connection info from environment variables set by Docker Compose.

#### Scenario: Fixtures connect to Docker services
- **WHEN** a test calls `fixture.NewTestTenant(t)` inside the test-runner container
- **THEN** it connects to MySQL using `DB_HOST` env var (Docker DNS: `mysql`) and creates a tenant database

#### Scenario: WireMock client connects to Docker service
- **WHEN** a test creates a WireMock client
- **THEN** it connects using `WIREMOCK_URL` env var (Docker DNS: `http://wiremock:8080`)

### Requirement: JWT token generation for tests
The system SHALL provide a `fixture.GenerateToken(AuthConfig)` function that creates valid JWT Bearer tokens using a test JWT secret.

#### Scenario: Generate cashier token
- **WHEN** a test calls `GenerateToken(AuthConfig{Source: "cashier", CompanyUuid: 123, StaffUuid: 456})`
- **THEN** it returns a `"Bearer <jwt>"` string that the service's auth middleware accepts

### Requirement: Test tenant lifecycle
The system SHALL provide a `fixture.NewTestTenant(tb)` function that creates an isolated MySQL database per test and cleans it up automatically.

#### Scenario: Create isolated tenant
- **WHEN** a test calls `fixture.NewTestTenant(t)`
- **THEN** it creates a `shop{uuid}` database with a unique UUID, runs GORM AutoMigrate for essential models, and registers a `t.Cleanup` that drops the database

#### Scenario: Parallel tests have separate databases
- **WHEN** two parallel tests each call `fixture.NewTestTenant(t)`
- **THEN** each receives a TestTenant with a different CompanyUuid and separate database

### Requirement: Data seeding helpers
The system SHALL provide typed seed functions (`SeedStaff`, `SeedDesk`, `SeedDevice`) that accept `testing.TB` and an override function pattern.

#### Scenario: Seed with defaults
- **WHEN** a test calls `fixture.SeedStaff(t, db)`
- **THEN** a Staff record with sensible defaults is inserted into the database and returned

#### Scenario: Seed with overrides
- **WHEN** a test calls `fixture.SeedStaff(t, db, func(s *model.Staff) { s.Name = "Custom" })`
- **THEN** a Staff record is inserted with the overridden Name field and other fields set to defaults

### Requirement: All fixtures use testing.TB interface
All fixture functions SHALL accept `testing.TB` (not `*testing.T`) to support reuse in both `Test*` and `Benchmark*` functions.

#### Scenario: Fixture works in benchmark
- **WHEN** a `Benchmark*` function calls `fixture.NewTestTenant(b)` where `b` is `*testing.B`
- **THEN** the function succeeds and `b.Cleanup()` is registered for teardown
