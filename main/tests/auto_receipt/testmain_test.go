//go:build integration

package auto_receipt_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

func TestMain(m *testing.M) {
	// auto_receipt tests are synchronous request/response flows. Drop tenant DBs
	// as soon as each test finishes so this heavy suite does not accumulate full
	// cloned schemas and exhaust the MySQL tmpfs used in CI.
	_ = os.Setenv("TEST_DROP_TENANT_DB", "1")

	if os.Getenv("TEST_ENABLE_RESOURCE_BUDGET") != "1" {
		os.Exit(m.Run())
	}

	release, err := fixture.AcquireResourceBudgetFromEnv(
		"full-tenant-heavy",
		1,
		10*time.Minute,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to acquire integration resource budget for auto_receipt: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	release()
	os.Exit(code)
}
