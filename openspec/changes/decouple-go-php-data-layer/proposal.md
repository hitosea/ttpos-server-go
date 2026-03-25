**Type:** Technical Debt

## Why

Go Main integration tests are tightly coupled to PHP Admin's database initialization scripts (`admin/database/seeds/saas.sql`, `shop_01.sql`). This coupling prevents Go Main from evolving independently, causes test failures due to SQL timing issues, and makes it difficult to eventually deprecate the legacy PHP Admin module.

## What Changes

- **BREAKING**: Remove dependency on `admin/database/seeds/saas.sql` for integration tests
- Create minimal test schema at `main/tests/fixtures/schema/saas_minimal.sql` containing only essential tables:
  - `ttpos_company` - Company/Group table for multi-tenancy
  - `ttpos_company_setting` - Company configuration
  - Minimal test data for integration testing
- Update `main/tests/docker-compose.yml` to use new minimal schema
- Keep existing PHP Admin seed files unchanged for backward compatibility

## Capabilities

### New Capabilities
- `integration-test-schema`: Minimal database schema for Go Main integration tests, decoupled from PHP Admin

### Modified Capabilities
- None (this is a test infrastructure change, no production behavior changes)

## Impact

**Affected Code:**
- `main/tests/docker-compose.yml` - Update db-init volume path and configuration
- `main/tests/` - New fixtures directory with test schema

**Dependencies:**
- Removes: Test dependency on `admin/database/seeds/` directory
- Adds: None (self-contained test schema)

**Systems:**
- Integration tests become self-contained and faster
- No changes to production code or APIs
