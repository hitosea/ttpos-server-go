---
name: integration-test
description: Write blackbox API integration tests for Main or BMP modules. Trigger when user asks to create, add, or write integration tests, or mentions testing an API endpoint. These are API-oriented blackbox tests — the test runner is a separate container with no internal imports.
allowed-tools: Read, Glob, Grep, Write, Edit, TodoWrite, AskUserQuestion, Bash
---

# Integration Test Writing Skill

## Overview

This skill guides writing blackbox API integration tests following the project convention.
Tests are organized **by business scenario** (file), with test cases **grouped by priority** (P0 → P1 → P2) within each file.

Full naming convention and rules: see [rules.md](rules.md)
Templates: [templates/main.go.tmpl](templates/main.go.tmpl), [templates/bmp.go.tmpl](templates/bmp.go.tmpl)

---

## Step 1 — Identify Target (MANDATORY)

Before asking the user anything, auto-detect what you can from their message:
- Endpoint URL → parse Terminal, Domain, Action
- Service name ("erp", "takeout", "message") → BMP module
- "cashier", "shop", "tablet" etc. → Main module terminal

Then conduct **Round 1** interview:

```yaml
AskUserQuestion:
  Q1: Which module?
    options: [Main (Gin/Go), BMP - ERP, BMP - Takeout, BMP - Message]
  Q2: What is the business scenario / feature being tested?
    (free text or paste the endpoint URL)
```

From the answers, derive:
- **Module**: main or bmp/{service}
- **Terminal**: Cashier | Shop | Tablet | Assistant | Kitchen | Kiosk | H5 | Member | Menu | Callboard | Erp | Takeout | Message
- **Domain**: Order | Auth | Product | Desk | Staff | Payment | Member | Setting | Stock | Health | Item | Template | Send | ...
- **Action**: Create | List | Get | Update | Delete | Sync | Push | Toggle | Login | Check | Send | ...

---

## Step 2 — Discover the API

Read the relevant files to understand the endpoint's request/response:

**For Main module:**
1. Read `main/router/router.go` to find the route group
2. Read the handler file in `main/app/api/v1/{terminal}/`
3. Read the request type from `main/app/dto/req/` for payload structure

**For BMP module:**
1. Read `ttpos-bmp/app/ttpos-{service}/internal/controller/` for HTTP routes
2. Read the protobuf in `ttpos-bmp/app/ttpos-{service}/api/` for gRPC
3. Read `ttpos-bmp/app/ttpos-{service}/internal/model/dto/` for request structure

---

## Step 3 — Scope Scenarios (MANDATORY)

Conduct **Round 2** interview:

```yaml
AskUserQuestion:
  Q1: Which priority levels to cover?
    options:
      - P0 only (critical path — release gate)
      - P0 + P1 (core + validation)
      - All: P0 + P1 + P2 (full coverage)
  Q2: Which scenario categories? (multiSelect)
    options:
      - HappyPath (P0 — success flow)
      - Unauthorized / AccessDenied (P0 — auth boundary)
      - InvalidInput / MissingField (P1 — validation)
      - NotFound (P1 — resource absence)
      - DuplicateEntry / Conflict (P1 — business rules)
      - Edge cases: EmptyCart, ZeroAmount, etc. (P2)
```

**Smart defaults by HTTP method** (pre-select if user doesn't specify):
- POST → P0: HappyPath, Unauthorized; P1: InvalidInput
- GET  → P0: HappyPath, Unauthorized; P1: NotFound
- DELETE → P0: HappyPath, Unauthorized; P1: NotFound

---

## Step 4 — Confirm Test Plan (MANDATORY)

Conduct **Round 3** — show the user exactly what will be generated:

```
Target file: main/tests/order/instant_order_test.go
Package: order_test

Tests to generate:
  P0  Test_P0_Cashier_InstantOrder_Create_HappyPath
  P0  Test_P0_Cashier_InstantOrder_Create_Unauthorized
  P1  Test_P1_Cashier_InstantOrder_Create_InvalidInput

Fixtures needed (Main):
  - NewTestTenantFull (full schema with shop_01.sql)
  - SeedCompany, SeedCompanySetting
  - SeedStaff (with WithStaffIsSuper, WithStaffDutyNo)
  - SeedDevice (with WithDeviceType, WithDeviceId)
  - GenerateStaffToken
  - WireMock: client.RegisterTTPOSDefaults

Fixtures needed (BMP):
  - fixture.NewErpClient() / NewTakeoutClient() / NewMessageClient()
```

Ask: "Does this look right? Proceed?"

---

## Step 5 — Generate the Test File

Check if the target file already exists:
- **Exists**: append new test functions in the correct priority section
- **New file**: create from the appropriate template

### Convention Checklist (enforce before writing)

1. ✅ File name: `{scenario}_test.go` (business scenario noun, snake_case)
2. ✅ Build tag: `//go:build integration` on line 1
3. ✅ Package: `{directory}_test` (external test package)
4. ✅ Function name pattern: `Test_{P[012]}_{Terminal}_{Domain}_{Action}_{Scenario}`
5. ✅ Tests grouped by priority with comment separators: `// --- P0 ---`, `// --- P1 ---`, `// --- P2 ---`
6. ✅ API paths as named constants (never inline strings)
7. ✅ Each test has a doc comment with `// Route:` annotation
8. ✅ No shared state between tests — each creates its own tenant/data
9. ✅ Cleanup via `t.Cleanup()` in fixture functions (not manual)

### File structure (from template):

```
// build tag
// package
// imports
// Constants block (paths, error codes)
// --- P0 ---   critical tests
// --- P1 ---   important tests
// --- P2 ---   edge case tests
// Helpers (file-local only)
```

---

## Step 6 — Run and Verify

After writing, run the new test(s) to verify they pass:

**Main:**
```bash
make test-main-local BUILD_ID=verify
```

**BMP:**
```bash
make test-bmp-local BUILD_ID=verify
```

---

## Quality Analysis Reference

```bash
# Run only P0 tests (release gate)
go test -tags=integration -run "^Test_P0_" ./...

# Run P0 + P1 (full gate)
go test -tags=integration -run "^Test_P[01]_" ./...

# Filter by terminal
go test -tags=integration -run "_Cashier_" ./...

# Filter by scenario type
go test -tags=integration -run "_HappyPath$" ./...

# Priority distribution report
go test -v -tags=integration ./... 2>&1 | grep -oP 'Test_P\d' | sort | uniq -c

# Scenario distribution report
go test -v -tags=integration ./... 2>&1 | grep -oP 'Test_\w+' | awk -F_ '{print $NF}' | sort | uniq -c | sort -rn
```
