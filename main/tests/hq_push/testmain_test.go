//go:build integration

package hq_push_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"ttpos-server-go/tests/fixture"
)

func TestMain(m *testing.M) {
	if os.Getenv("TEST_ENABLE_RESOURCE_BUDGET") != "1" {
		os.Exit(m.Run())
	}

	release, err := fixture.AcquireResourceBudgetFromEnv(
		"full-tenant-heavy",
		1,
		10*time.Minute,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to acquire integration resource budget for hq_push: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	release()
	os.Exit(code)
}
