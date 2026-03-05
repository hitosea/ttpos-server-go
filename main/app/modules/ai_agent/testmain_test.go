package ai_agent

import (
	"os"
	"testing"

	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Initialize a no-op logger to prevent nil pointer panics in tests
	logger.Logger = zap.NewNop()
	os.Exit(m.Run())
}
