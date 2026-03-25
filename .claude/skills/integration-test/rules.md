# Integration Test Naming Convention & Rules

## Test ID Structure

```
Test_{Priority}_{Terminal}_{Domain}_{Action}_{Scenario}
```

### Priority (P0 / P1 / P2)

| Level | Meaning | Release gate? |
|-------|---------|--------------|
| **P0** | Critical business path — must pass to release | ✅ Blocks release |
| **P1** | Important validation & error handling | ⚠️ Warns on failure |
| **P2** | Edge cases & boundary conditions | ℹ️ Informational only |

### Terminal

Maps to the API group (Main) or BMP service:

| Terminal | Source | Example path |
|----------|--------|--------------|
| `Cashier` | Main | `/api/v1/cashier/...` |
| `Shop` | Main | `/api/v1/shop/...` |
| `Tablet` | Main | `/api/v1/tablet/...` |
| `Assistant` | Main | `/api/v1/assistant/...` |
| `Kitchen` | Main | `/api/v1/kitchen/...` |
| `Kiosk` | Main | `/api/v1/kiosk/...` |
| `H5` | Main | `/api/v1/h5/...` |
| `Member` | Main | `/api/v1/member/...` |
| `Menu` | Main | `/api/v1/menu/...` |
| `Callboard` | Main | `/api/v1/callboard/...` |
| `Erp` | BMP | `bmp-erp:14021` |
| `Takeout` | BMP | `bmp-takeout:14031` |
| `Message` | BMP | `bmp-message:14041` |

### Domain

Business domain, PascalCase. Examples:

`Order` · `Auth` · `Product` · `Desk` · `Staff` · `Payment` · `Member` · `Setting` · `Stock` · `Health` · `Item` · `Template` · `Notification` · `Callback`

Compound domains: `InstantOrder` · `DineInOrder` · `BuffetOrder` · `OrderPayment`

### Action

The operation being tested. Use the HTTP method intent:

`Create` · `List` · `Get` · `Update` · `Delete` · `Login` · `Logout` · `Sync` · `Push` · `Toggle` · `Send` · `Check`

### Scenario

What behavior is being validated:

| Category | Scenario suffixes | Default priority |
|----------|------------------|-----------------|
| Success | `HappyPath` | P0 |
| Auth | `Unauthorized`, `AccessDenied`, `Expired`, `TokenInvalid` | P0 |
| Validation | `InvalidInput`, `MissingField`, `EmptyBody`, `InvalidFormat` | P1 |
| Not found | `NotFound` | P1 |
| Business rules | `DuplicateEntry`, `Conflict`, `LimitExceeded`, `AlreadyExists` | P1 |
| Edge cases | `EmptyCart`, `ZeroAmount`, `ConcurrentAccess`, `MaxItems` | P2 |

---

## File Organization

```
{module}/tests/{service_or_domain}/{business_scenario}_test.go
```

### File naming examples

```
main/tests/
  order/
    instant_order_test.go      # instant order creation/cart lifecycle
    dine_in_order_test.go      # dine-in with desk assignment
    buffet_order_test.go       # buffet ordering
    order_payment_test.go      # payment flows for orders
  auth/
    cashier_login_test.go      # cashier terminal login
    device_binding_test.go     # device bind/unbind
  product/
    product_crud_test.go       # product create/read/update/delete
    sold_out_test.go           # sold-out toggling
  desk/
    desk_assignment_test.go    # table assignment and release

ttpos-bmp/tests/
  erp/
    item_sync_test.go          # ERP item synchronization
    stock_test.go              # stock operations
  takeout/
    order_push_test.go         # push order to takeout providers
    grab_callback_test.go      # Grab webhook callbacks
  message/
    email_notification_test.go # email sending via templates
    template_test.go           # template CRUD
```

Rule: **one file per distinct business scenario**, not one file per endpoint.

---

## Test File Structure

```go
//go:build integration

package {directory}_test  // external test package (blackbox)

import (
    "testing"
    "ttpos-server-go/tests/fixture"  // Main
    // OR
    "ttpos-bmp/tests/fixture"        // BMP
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
    // API paths — never use inline strings in test calls
    pathInstantOrderCreate = "/api/v1/cashier/instant/order/create"
    pathInstantOrderCart   = "/api/v1/cashier/instant/order/cart/info"

    // Error codes matching server-side constants
    codeTokenInvalid = -102  // constant.CodeTokenInvalid
    codeAccessDenied = -103  // constant.CodeAccessDenied
)

// ---------------------------------------------------------------------------
// P0 — Critical Path
// ---------------------------------------------------------------------------

// Test_P0_Cashier_InstantOrder_Create_HappyPath tests the complete instant
// order creation workflow.
// Route: POST /api/v1/cashier/instant/order/create
func Test_P0_Cashier_InstantOrder_Create_HappyPath(t *testing.T) {
    // ...
}

// Test_P0_Cashier_InstantOrder_Create_Unauthorized tests that requests
// without a token are rejected with codeTokenInvalid.
// Route: POST /api/v1/cashier/instant/order/create
func Test_P0_Cashier_InstantOrder_Create_Unauthorized(t *testing.T) {
    // ...
}

// ---------------------------------------------------------------------------
// P1 — Important Validation
// ---------------------------------------------------------------------------

// Test_P1_Cashier_InstantOrder_GetCart_AccessDenied tests that a valid JWT
// for a company without seeded data is rejected with codeAccessDenied.
// Route: GET /api/v1/cashier/instant/order/cart/info
func Test_P1_Cashier_InstantOrder_GetCart_AccessDenied(t *testing.T) {
    // ...
}

// ---------------------------------------------------------------------------
// P2 — Edge Cases
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
```

---

## Rules

### Preflight contract check (mandatory before writing assertions)

1. Read the handler/controller, request DTO, and at least one service/usecase layer.
2. Open one neighboring integration test in the same directory and follow its fixture style.
3. Write down every request UUID field and map it to the seeded source object.
4. If a fixture helper does not expose a needed UUID, extend the helper result in test code instead of guessing or hardcoding.
5. Check whether the API path triggers async work after the response returns; if yes, avoid manual teardown in the test body.

### Mandatory
1. `//go:build integration` must be the first line
2. Package must be `{directory}_test` — external (blackbox)
3. API paths must be named constants, never inline strings
4. Each test function must have a doc comment with `// Route:` annotation
5. Tests must be grouped by priority (P0 → P1 → P2) with comment separators
6. No shared state between tests — each test is fully self-contained
7. Cleanup via `t.Cleanup()` in fixture functions only

### Known traps

- **Package requests**: when the product is a package (`product_type = 1`), verify whether the top-level item also needs `product_package_uuid`. Do not assume `flavor_uuid` alone is enough.
- **UUID semantics**: `product_package_uuid` (parent package) and `flavor_uuid` / BOM UUID (spec) are different values. Never reuse one as the other.
- **Main integration module root**: `main/tests` has its own `go.mod`; run focused checks from `main/tests` or via `make test-main-local`.
- **Async cleanup**: do not manually drop tenant DBs / add force cleanup in test bodies for async scenarios such as HQ push.

### Forbidden
- ❌ Importing any `ttpos-bmp/` or `ttpos-server-go/app/` packages (blackbox only)
- ❌ Inline API path strings (use named constants)
- ❌ Shared mutable state across tests (no `var` at package level)
- ❌ Manual `t.Cleanup()` calls — fixtures handle this
- ❌ Hardcoding `0` / guessed UUIDs just to satisfy request fields
- ❌ Changing production code first when the failure can be fixed by correcting test seed data or request construction

### Fixture availability

**Main tests** (`main/tests/fixture/`):
- `fixture.NewHTTPClient()` — HTTP client with fluent assertions
- `fixture.NewTestTenant(t, uuid)` — minimal schema (fast)
- `fixture.NewTestTenantFull(t, uuid)` — full schema via shop_01.sql
- `fixture.GenerateStaffToken(t, companyUUID, ...)` — JWT generation
- `fixture.SeedStaff(t, db, ...opts)` — seed staff with options
- `fixture.SeedDevice(t, db, ...opts)` — seed device
- `fixture.SeedCompany(t, db, ...opts)` — seed company
- `fixture.SeedProduct(t, db, ...opts)` — seed product
- `fixture.GenerateCompanyUUID(t)` — unique company UUID

**BMP tests** (`ttpos-bmp/tests/fixture/`):
- `fixture.NewErpClient()` — HTTP client → bmp-erp:14021
- `fixture.NewTakeoutClient()` — HTTP client → bmp-takeout:14031
- `fixture.NewMessageClient()` — HTTP client → bmp-message:14041

---

## Quality Analysis Commands

```bash
# Release gate: P0 must pass 100%
go test -tags=integration -run "^Test_P0_" ./...

# Full gate: P0 + P1
go test -tags=integration -run "^Test_P[01]_" ./...

# By service / terminal
go test -tags=integration -run "_Cashier_" ./...
go test -tags=integration -run "_Erp_" ./...

# By scenario type
go test -tags=integration -run "_HappyPath$" ./...
go test -tags=integration -run "_Unauthorized$" ./...

# Priority distribution (version-end report)
go test -v -tags=integration ./... 2>&1 | grep -oP 'Test_P\d' | sort | uniq -c
#   15 Test_P0
#   10 Test_P1
#    5 Test_P2

# Scenario type distribution
go test -v -tags=integration ./... 2>&1 \
  | grep -oP '--- (PASS|FAIL): Test_\w+' \
  | awk -F_ '{print $NF}' \
  | sort | uniq -c | sort -rn
```
