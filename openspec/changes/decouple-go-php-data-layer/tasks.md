## 1. Create Test Schema Directory Structure

- [ ] 1.1 Create `main/tests/fixtures/schema/` directory
- [ ] 1.2 Add README explaining the fixtures directory purpose

## 2. Create Minimal SAAS Schema

- [ ] 2.1 Create `main/tests/fixtures/schema/saas_minimal.sql` with file header documentation
- [ ] 2.2 Add `ttpos_company` table definition matching production schema
- [ ] 2.3 Add `ttpos_company_setting` table definition matching production schema
- [ ] 2.4 Add test company INSERT statement (uuid: 1000000000000001, status: 1)
- [ ] 2.5 Add test company_setting INSERT statement linked to test company

## 3. Fix Docker Compose Configuration

- [ ] 3.1 Update db-init volume mount from `../../admin/database/seeds` to `../../main/tests/fixtures/schema`
- [ ] 3.2 Convert db-init command from YAML folding to JSON array syntax
- [ ] 3.3 Update SQL file reference from `saas.sql` to `saas_minimal.sql`
- [ ] 3.4 Simplify environment variables (use hardcoded test values where appropriate)

## 4. Verify Integration Tests

- [ ] 4.1 Run `make test-clean` to ensure clean state
- [ ] 4.2 Run `make test-local` to verify tests work with new schema
- [ ] 4.3 Confirm no "syntax error near unexpected token" errors in db-init
- [ ] 4.4 Verify MySQL health check passes with new configuration

## 5. Documentation

- [ ] 5.1 Update main/tests/README.md (if exists) to document new fixtures location
- [ ] 5.2 Add comment in saas_minimal.sql referencing production schema source
- [ ] 5.3 Document how to extend the schema for additional test tables
