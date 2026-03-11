## ADDED Requirements

### Requirement: Unified Main service compose file with profiles
The system SHALL provide a single `docker-compose.yml` at the repo root that replaces `docker-compose.dev.yml`, `docker-compose.dev.redis.yml`, and `docker-compose.production.yml`. It SHALL support `dev`, `test`, and `prod` profiles via Docker Compose v2 `--profile` flag.

#### Scenario: Start development environment
- **WHEN** user runs `docker compose --profile dev up -d`
- **THEN** the system starts DB (MySQL 8.0), Redis, PHP, Nginx, Golang dev, and reconciliation_templates services

#### Scenario: Start test infrastructure
- **WHEN** user runs `docker compose --profile test up -d`
- **THEN** the system starts DB (MySQL 8.0), WireMock, ttpos-server-go (production image), and test-runner containers. No ports are exposed to host — all communication via internal Docker network.

#### Scenario: Start production environment
- **WHEN** user runs `docker compose --profile prod up -d`
- **THEN** the system starts DB (configurable image via `DB_IMAGE`), Redis, PHP, Nginx (production image), Golang (production image), reconciliation_templates, and envoy-mirror services

### Requirement: Unified BMP service compose file with profiles
The system SHALL provide a single `ttpos-bmp/docker-compose.yml` that replaces `docker-compose.yml`, `docker-compose.dev.yml`, `docker-compose.mid.yml`, and `docker-compose.production.yml` in the ttpos-bmp directory. It SHALL support `mid`, `dev`, and `prod` profiles.

#### Scenario: Start BMP middleware only
- **WHEN** user runs `docker compose --profile mid up -d` in ttpos-bmp/
- **THEN** the system starts Nacos, RocketMQ (namesrv + broker), and Jaeger services

#### Scenario: Start full BMP development stack
- **WHEN** user runs `docker compose --profile mid --profile dev up -d` in ttpos-bmp/
- **THEN** the system starts middleware services AND BMP microservices (ERP, Takeout, Message, WebSocket) built from local Dockerfiles

#### Scenario: Start BMP production
- **WHEN** user runs `docker compose --profile prod up -d` in ttpos-bmp/
- **THEN** the system starts BMP microservices using pre-built images from `hub.hitosea.com` tagged with `${IMAGE_VERSION}`

### Requirement: Test profile with isolated network
The test profile SHALL run all services (mysql, wiremock, ttpos-server-go, test-runner) in an internal Docker network with NO port mapping to the host. Communication SHALL use Docker DNS container names.

#### Scenario: Services communicate via Docker DNS
- **WHEN** the test profile is running
- **THEN** ttpos-server-go connects to `mysql:3306` and `wiremock:8443`, and test-runner connects to `ttpos-server-go:8080` and `wiremock:8080`

#### Scenario: No host ports exposed
- **WHEN** the test profile is running
- **THEN** no container ports are mapped to localhost — enabling multiple parallel test runs without port conflicts

### Requirement: WireMock container in test profile
The system SHALL include a WireMock 3.x container in the `test` profile that supports both HTTP stub matching and gRPC stub matching via the WireMock gRPC extension.

#### Scenario: WireMock serves HTTP stubs
- **WHEN** the test profile is running and test-runner registers a stub via `http://wiremock:8080/__admin/mappings`
- **THEN** HTTP requests matching the stub pattern receive the configured response

#### Scenario: WireMock serves gRPC stubs
- **WHEN** the test profile is running with proto descriptors mounted and a gRPC stub is registered
- **THEN** gRPC calls from ttpos-server-go to `wiremock:8443` matching the stub pattern receive the configured protobuf response

### Requirement: Test-runner container
The test profile SHALL include a `test-runner` container that runs `go test -tags=integration` and exits with the test result code.

#### Scenario: Test-runner executes tests
- **WHEN** the test profile starts with `docker compose --profile test up --abort-on-container-exit`
- **THEN** test-runner container runs the integration tests and the compose command exits with the test result code

#### Scenario: Test-runner connects to services
- **WHEN** test-runner starts
- **THEN** it has environment variables `SERVICE_URL=http://ttpos-server-go:8080`, `DB_HOST=mysql`, `WIREMOCK_URL=http://wiremock:8080`

### Requirement: Docker Compose project isolation
Each test run SHALL use a unique project name (`-p test-$BUILD_ID`) to enable parallel test runs with complete network and container isolation.

#### Scenario: Parallel test runs are isolated
- **WHEN** CI runs `docker compose -p test-123 --profile test up` and `docker compose -p test-456 --profile test up` simultaneously
- **THEN** both runs have separate Docker networks, containers, and volumes with zero conflicts

### Requirement: Environment file separation
The system SHALL use `.env` for shared defaults and `.env.test` for test-specific overrides (JWT secret, DB credentials, WireMock config).

#### Scenario: Test profile uses test env file
- **WHEN** user runs `docker compose --profile test --env-file .env.test up -d`
- **THEN** services use test-specific configuration from `.env.test`

### Requirement: No cross-references between compose files
The Main compose file SHALL NOT reference any BMP services or networks. The BMP compose file SHALL NOT reference any Main services or networks. Each file SHALL be self-contained for future repository separation.

#### Scenario: Main compose is independent
- **WHEN** the ttpos-bmp/ directory is removed
- **THEN** `docker compose --profile dev up -d` at repo root still succeeds without errors
