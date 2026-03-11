## Context

**Current State:**
- Integration tests in `main/tests/` use `admin/database/seeds/saas.sql` for database initialization
- The saas.sql file contains ~500+ lines with PHP Admin specific tables (`ttpos_admin_*`, `ttpos_client_*`, etc.)
- db-init container loads this SQL via docker-compose volume mount: `../../admin/database/seeds:/docker-entrypoint-initdb.d`
- YAML folding syntax (`>`) in docker-compose causes shell script syntax errors in db-init command

**Constraints:**
- PHP Admin seed files cannot be modified (legacy system still in production)
- Production database schema must remain unchanged
- Integration tests need only `ttpos_company` and `ttpos_company_setting` tables to run

**Stakeholders:**
- Go Main developers: Faster, more reliable integration tests
- PHP Admin maintainers: No impact on their workflow
- DevOps: Simpler test infrastructure

## Goals / Non-Goals

**Goals:**
- Create minimal test schema with only essential tables for Go Main integration tests
- Remove test dependency on PHP Admin's seed files
- Fix docker-compose YAML syntax issues in db-init command
- Make integration tests self-contained and faster to initialize

**Non-Goals:**
- Modifying PHP Admin's seed files
- Changes to production database schema
- Migrating PHP Admin functionality to Go (out of scope)

## Decisions

### 1. Minimal Schema Approach
**Decision:** Create `saas_minimal.sql` with only `ttpos_company` and `ttpos_company_setting` tables plus test data.

**Rationale:**
- Integration tests only need company configuration to initialize the application
- Smaller SQL = faster container startup
- Go Main can evolve its test schema independently

**Alternatives considered:**
- *Use existing saas.sql*: Rejected due to coupling and timing issues
- *Empty schema*: Rejected, tests need at least one company record
- *ORM-based seeding*: Rejected, adds complexity and test execution time

### 2. JSON Array Syntax for Docker Commands
**Decision:** Use JSON array syntax for shell commands in docker-compose.yml instead of YAML folding.

**Rationale:**
- YAML folding (`>`) removes newlines, breaking multi-line shell scripts
- JSON array `["cmd", "arg1", "arg2"]` preserves command structure
- More explicit and less error-prone

**Example:**
```yaml
# Before (broken):
command: >
  sh -c "for i in 1..10; do echo $i; done"

# After (fixed):
command: ["sh", "-c", "for i in 1..10; do echo $i; done"]
```

### 3. Fixtures Directory Structure
**Decision:** Place test schemas under `main/tests/fixtures/schema/`.

**Rationale:**
- Clear separation: test fixtures vs production code
- Follows common patterns (`fixtures/` for test data)
- Easy to extend with merchant-specific schemas later

## Risks / Trade-offs

### Schema Drift
**Risk:** Test schema may diverge from production schema over time, causing tests to pass while production fails.

**Mitigation:**
- Document that test schema is MINIMAL only for initialization
- Add comment in saas_minimal.sql referencing source tables
- Consider adding schema validation step in CI

### Missing Tables for Future Tests
**Risk:** Future integration tests may need additional tables not in minimal schema.

**Mitigation:**
- Easy to extend: just add tables to saas_minimal.sql
- Schema is version controlled, changes are traceable
- Each test can specify required tables in doc comments

## Migration Plan

1. **Create minimal schema:** `main/tests/fixtures/schema/saas_minimal.sql`
2. **Update docker-compose.yml:** Change volume mount and fix command syntax
3. **Verify tests run:** Execute `make test-local` to confirm
4. **Update documentation:** Add README in fixtures/ directory

**Rollback:** Revert docker-compose.yml changes to use `admin/database/seeds/` path.

## Open Questions

None - this is a straightforward infrastructure improvement with clear scope.
