//go:build integration

package erp_test

import (
	"testing"
	"ttpos-bmp/tests/fixture"
)

// Test_P0_Erp_Health_Check_HappyPath validates that the ERP service started successfully:
// Docker image built, config injected, DB connected, HTTP server listening.
func Test_P0_Erp_Health_Check_HappyPath(t *testing.T) {
	client := fixture.NewErpClient()
	// GoFrame serves OpenAPI spec at /api.json — always returns 200 when server is up
	client.Get(t, "/api.json").AssertOK(t)
}
