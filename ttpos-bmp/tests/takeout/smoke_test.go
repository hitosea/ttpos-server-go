//go:build integration

package takeout_test

import (
	"testing"
	"ttpos-bmp/tests/fixture"
)

// Test_P0_Takeout_Health_Check_HappyPath validates that the Takeout service started successfully:
// Docker image built, config injected, DB connected, HTTP server listening.
func Test_P0_Takeout_Health_Check_HappyPath(t *testing.T) {
	client := fixture.NewTakeoutClient()
	// GoFrame serves OpenAPI spec at /api.json — always returns 200 when server is up
	client.Get(t, "/api.json").AssertOK(t)
}
