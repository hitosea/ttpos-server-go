# Unit Test Rules & Best Practices

## Test Categories

### Category 1: Pure Tests (no build tag)

Tests that need **zero external services**. These run in CI, locally, and everywhere.

```go
package utils_test  // or package utils (for internal tests)

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestParseAmount(t *testing.T) {
    // pure logic — no Redis, no MySQL, no network
}
```

**No build tag required.** These tests MUST pass with just `go test ./...`.

### Category 2: Service-Dependent Tests (build tag required)

Tests that require Redis, MySQL, gRPC, or other running services.

```go
//go:build integration

package cache

import (
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    // shared setup (logger, etc.)
    os.Exit(m.Run())
}

func TestRedisCache_Set(t *testing.T) {
    host := os.Getenv("REDIS_HOST")
    if host == "" {
        t.Skip("REDIS_HOST not set — skipping Redis test")
    }
    // ...
}
```

**Rules:**
1. MUST have `//go:build integration` on line 1
2. MUST read connection info from environment variables
3. MUST `t.Skip()` when the required service env var is empty

---

## Environment Variable Convention

### Redis
```go
host     := envOrDefault("REDIS_HOST", "")        // empty = skip
port     := envOrDefault("REDIS_PORT", "6379")
password := os.Getenv("REDIS_PASSWORD")            // empty = no auth
db       := envOrDefault("REDIS_DB", "0")
```

### MySQL
```go
host     := envOrDefault("DB_HOST", "")            // empty = skip
port     := envOrDefault("DB_PORT", "3306")
user     := envOrDefault("DB_USERNAME", "root")
password := os.Getenv("DB_PASSWORD")
database := envOrDefault("DB_DATABASE", "test")
```

### gRPC
```go
addr        := envOrDefault("GRPC_ADDR", "")       // empty = skip
tlsInsecure := os.Getenv("GRPC_TLS_INSECURE")     // "true" for test envs
```

### Helper function (copy into test file if needed)
```go
func envOrDefault(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

## Table-Driven Test Pattern

Use for any function with multiple input/output combinations:

```go
func TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        rate     float64
        expected float64
    }{
        {"zero amount", 0, 0.1, 0},
        {"standard discount", 100, 0.1, 10},
        {"full discount", 100, 1.0, 100},
        {"no discount", 100, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateDiscount(tt.amount, tt.rate)
            assert.InDelta(t, tt.expected, result, 0.001)
        })
    }
}
```

---

## TestMain Pattern

Use when the package needs shared initialization (logger, connection pools):

```go
func TestMain(m *testing.M) {
    // Initialize shared dependencies
    logger.Logger = zap.NewNop()  // silent logger for tests

    // Optional: setup/teardown
    code := m.Run()

    // Cleanup if needed
    os.Exit(code)
}
```

**When to use TestMain:**
- Package uses `logger.Logger` (will panic if nil)
- Package needs a one-time connection pool
- Package needs to set up test fixtures shared across all tests in the file

**When NOT to use TestMain:**
- Pure logic tests with no shared state
- Tests that are fully self-contained

---

## Test Naming Convention

### Pure tests
```
TestFunctionName                    # single scenario
TestFunctionName_ScenarioName       # specific scenario
TestFunctionName_EdgeCases          # boundary conditions
BenchmarkFunctionName               # performance
```

### Service-dependent tests
```
TestRedisCache_Set                  # component + action
TestRedisCache_Set_Expiry           # component + action + scenario
TestMySQLRepo_FindByUUID_NotFound   # component + action + scenario
```

---

## Assertion Style

### Prefer testify/assert
```go
// ✅ Good — clear, concise, descriptive failure messages
assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.Nil(t, result)
assert.Contains(t, str, "substring")
assert.Len(t, slice, 3)

// ✅ Also good for fatal assertions (stop test immediately)
require.NoError(t, err)  // use when subsequent lines depend on no error
```

### Acceptable — raw checks for complex assertions
```go
// ✅ OK for complex logic that testify doesn't express well
if len(results) != expectedCount {
    t.Fatalf("expected %d results, got %d", expectedCount, len(results))
}
```

### Forbidden
```go
// ❌ NEVER use panic in tests
if err != nil {
    panic(err)  // WRONG — use t.Fatal(err)
}
```

---

## Mocking Strategy

### Option 1: Interface-based mocking (preferred)
```go
type MockRepository struct {
    FindFunc func(uuid string) (*Model, error)
}

func (m *MockRepository) Find(uuid string) (*Model, error) {
    return m.FindFunc(uuid)
}
```

### Option 2: Functional injection
```go
func TestProcessOrder(t *testing.T) {
    // Inject a fake dependency via function field
    svc := &OrderService{
        lookupProduct: func(uuid string) (*Product, error) {
            return &Product{Name: "Test", Price: 10.00}, nil
        },
    }
    // ...
}
```

### Option 3: Real service with env vars (for service-dependent tests)
```go
func TestWithRealRedis(t *testing.T) {
    host := os.Getenv("REDIS_HOST")
    if host == "" {
        t.Skip("REDIS_HOST not set")
    }
    // Use real Redis — but NEVER hardcode the address
}
```

---

## File Organization

```
main/pkg/cache/
    group.go              # source code
    group_test.go          # pure tests (no build tag)
    group_redis_test.go    # Redis-dependent tests (//go:build integration)
    example_test.go        # example tests for godoc

main/app/service/
    order_service.go       # source code
    order_service_test.go  # unit tests with mocks (no build tag)
```

**Rule:** Separate pure tests from service-dependent tests into different files when both exist for the same package.

---

## Common Anti-Patterns (REJECT these)

### 1. Hardcoded service addresses
```go
// ❌ NEVER DO THIS
rdb := redis.NewClient(&redis.Options{
    Addr: "192.168.100.69:6379",     // hardcoded IP
    Password: "sass@123.com!",        // hardcoded credential
})

// ✅ CORRECT
host := os.Getenv("REDIS_HOST")
if host == "" {
    t.Skip("REDIS_HOST not set")
}
rdb := redis.NewClient(&redis.Options{
    Addr:     net.JoinHostPort(host, envOrDefault("REDIS_PORT", "6379")),
    Password: os.Getenv("REDIS_PASSWORD"),
})
```

### 2. Tests that crash when service is down
```go
// ❌ NEVER DO THIS — panics if Redis is unreachable
func TestSomething(t *testing.T) {
    rdb := redis.NewClient(...)
    err := rdb.Ping(ctx).Err()
    if err != nil {
        panic(err)  // kills entire test suite
    }
}

// ✅ CORRECT — graceful skip
func TestSomething(t *testing.T) {
    host := os.Getenv("REDIS_HOST")
    if host == "" {
        t.Skip("REDIS_HOST not set")
    }
    rdb := redis.NewClient(...)
    err := rdb.Ping(ctx).Err()
    require.NoError(t, err, "Redis connection failed")
}
```

### 3. Using panic instead of test assertions
```go
// ❌ NEVER
if err != nil {
    panic(err)
}

// ✅ CORRECT
require.NoError(t, err)
// or
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
```

### 4. Shared mutable state across tests
```go
// ❌ NEVER — tests pollute each other
var globalCounter int

func TestA(t *testing.T) { globalCounter++ }
func TestB(t *testing.T) { /* depends on globalCounter from TestA */ }

// ✅ CORRECT — each test is independent
func TestA(t *testing.T) {
    counter := 0
    counter++
    assert.Equal(t, 1, counter)
}
```
