package fixture

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	defaultResourceBudgetTimeout = 10 * time.Minute
	defaultResourceBudgetPoll    = 250 * time.Millisecond
)

// AcquireResourceBudget acquires one slot from a cross-process MySQL advisory-lock
// budget. Use it from TestMain of heavy integration suites so only a bounded number
// of packages can hold expensive resources at the same time.
func AcquireResourceBudget(name string, slots int, timeout time.Duration) (func(), error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("resource budget name cannot be empty")
	}
	if slots <= 0 {
		return nil, fmt.Errorf("resource budget %q must have at least one slot", name)
	}
	if timeout <= 0 {
		timeout = defaultResourceBudgetTimeout
	}

	rootConfig := RootDBConfig()
	lockDB, err := sql.Open("mysql", rootConfig.DSNWithoutDB())
	if err != nil {
		return nil, fmt.Errorf("open budget lock connection: %w", err)
	}
	configureAdminDBPool(lockDB)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	conn, err := lockDB.Conn(ctx)
	if err != nil {
		cancel()
		lockDB.Close()
		return nil, fmt.Errorf("reserve budget lock connection: %w", err)
	}

	deadline := time.Now().Add(timeout)
	pollInterval := time.Duration(getEnvInt("TEST_RESOURCE_BUDGET_POLL_MS", int(defaultResourceBudgetPoll/time.Millisecond))) * time.Millisecond
	if pollInterval <= 0 {
		pollInterval = defaultResourceBudgetPoll
	}

	var acquiredLock string
	for time.Now().Before(deadline) {
		for slot := 0; slot < slots; slot++ {
			lockName := fmt.Sprintf("ttpos_test_budget:%s:%d", sanitizeBudgetName(name), slot)
			var result sql.NullInt64
			if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 0)", lockName).Scan(&result); err != nil {
				conn.Close()
				cancel()
				lockDB.Close()
				return nil, fmt.Errorf("acquire budget lock %q: %w", lockName, err)
			}
			if result.Valid && result.Int64 == 1 {
				acquiredLock = lockName
				released := false
				release := func() {
					if released {
						return
					}
					released = true
					_, _ = conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", acquiredLock)
					_ = conn.Close()
					cancel()
					_ = lockDB.Close()
				}
				return release, nil
			}
		}

		select {
		case <-ctx.Done():
			conn.Close()
			cancel()
			lockDB.Close()
			return nil, fmt.Errorf("timeout waiting for resource budget %q (slots=%d)", name, slots)
		case <-time.After(pollInterval):
		}
	}

	conn.Close()
	cancel()
	lockDB.Close()
	return nil, fmt.Errorf("timeout waiting for resource budget %q (slots=%d)", name, slots)
}

// AcquireResourceBudgetFromEnv acquires a budget slot using per-budget env overrides:
//
//	TEST_RESOURCE_BUDGET_<NAME>_SLOTS
//	TEST_RESOURCE_BUDGET_<NAME>_TIMEOUT_SEC
func AcquireResourceBudgetFromEnv(name string, defaultSlots int, defaultTimeout time.Duration) (func(), error) {
	envKey := strings.ToUpper(sanitizeBudgetName(name))
	slots := getEnvInt(fmt.Sprintf("TEST_RESOURCE_BUDGET_%s_SLOTS", envKey), defaultSlots)
	timeoutSec := getEnvInt(fmt.Sprintf("TEST_RESOURCE_BUDGET_%s_TIMEOUT_SEC", envKey), int(defaultTimeout/time.Second))
	return AcquireResourceBudget(name, slots, time.Duration(timeoutSec)*time.Second)
}

func sanitizeBudgetName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_", "-", "_", ".", "_")
	return replacer.Replace(strings.ToLower(name))
}
