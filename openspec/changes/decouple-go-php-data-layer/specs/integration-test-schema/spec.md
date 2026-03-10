## ADDED Requirements

### Requirement: Minimal test schema provides essential company tables
The system SHALL provide a minimal database schema containing only the tables required for Go Main integration tests, independent of PHP Admin seed files.

The schema MUST include:
- `ttpos_company` table with company/group data structure
- `ttpos_company_setting` table with company configuration
- One test company record for integration testing

The schema MUST NOT include:
- PHP Admin specific tables (ttpos_admin_*, ttpos_client_*, etc.)
- Unnecessary tables for test initialization

#### Scenario: Schema file exists at expected path
- **WHEN** integration test framework initializes
- **THEN** system loads `main/tests/fixtures/schema/saas_minimal.sql`
- **AND** file contains only essential tables

#### Scenario: Schema creates test company
- **WHEN** schema is loaded into test database
- **THEN** `ttpos_company` table contains one test company
- **AND** `ttpos_company_setting` table contains corresponding settings
- **AND** company has `uuid = 1000000000000001`
- **AND** company status is enabled (status = 1)

### Requirement: Test schema decoupled from PHP Admin
The system SHALL NOT depend on PHP Admin's `admin/database/seeds/` directory for integration test initialization.

#### Scenario: Tests run without PHP Admin directory
- **WHEN** `admin/database/seeds/` directory is removed or inaccessible
- **THEN** integration tests still run successfully
- **AND** db-init container uses `main/tests/fixtures/schema/` instead

### Requirement: Docker command uses proper syntax
The db-init container command in docker-compose.yml SHALL use JSON array syntax to avoid YAML folding issues.

#### Scenario: Command executes without syntax errors
- **WHEN** docker-compose starts the db-init container
- **THEN** shell command parses correctly
- **AND** no "syntax error near unexpected token" errors occur
- **AND** SQL file loads successfully

#### Scenario: Retry logic handles MySQL startup timing
- **WHEN** MySQL is not immediately ready
- **THEN** db-init retries connection up to 30 times
- **AND** exits successfully once SQL loads
- **AND** exits with error after 30 failed attempts

### Requirement: Test schema documentation
The system SHALL include documentation explaining the minimal schema approach and how to extend it.

#### Scenario: Schema file is self-documenting
- **WHEN** developer opens `saas_minimal.sql`
- **THEN** file header explains its purpose
- **AND** comments indicate which tables are included and why
- **AND** references to production schema are provided
