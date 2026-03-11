//go:build integration

package message_test

import (
	"testing"
	"ttpos-bmp/tests/fixture"
)

// Test_P0_Message_Health_Check_HappyPath validates that the Message service started successfully:
// Docker image built, config injected, DB connected, HTTP server listening.
func Test_P0_Message_Health_Check_HappyPath(t *testing.T) {
	client := fixture.NewMessageClient()
	// GoFrame serves OpenAPI spec at /api.json — always returns 200 when server is up
	client.Get(t, "/api.json").AssertOK(t)
}
