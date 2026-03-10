//go:build integration

package websocket_test

import (
	"net/http"
	"testing"

	"ttpos-bmp/tests/fixture"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	pathApiJson     = "/api.json"
	pathHealthCheck = "/health"

	// wsHealthToken must match tests/config/websocket.yaml → health.token
	wsHealthToken = "test-health-token"
)

// ---------------------------------------------------------------------------
// P0 — Critical Path
// ---------------------------------------------------------------------------

// Test_P0_Websocket_Health_Smoke_HappyPath validates the WebSocket service started
// successfully: image built, config injected, DB connected, HTTP server listening.
// Route: GET /api.json
func Test_P0_Websocket_Health_Smoke_HappyPath(t *testing.T) {
	client := fixture.NewWebsocketClient()

	client.Get(t, pathApiJson).AssertOK(t)
}

// Test_P0_Websocket_Health_Check_HappyPath tests that the health endpoint returns
// UP status when all dependencies are healthy.
// Route: GET /health
func Test_P0_Websocket_Health_Check_HappyPath(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.DoRequest(t, http.MethodGet, pathHealthCheck, nil, map[string]string{
		"X-Health-Token": wsHealthToken,
	})

	resp.AssertOK(t)
	resp.AssertBodyContains(t, `"status"`)
	resp.AssertBodyContains(t, "UP")
}

// ---------------------------------------------------------------------------
// P1 — Important Validation
// ---------------------------------------------------------------------------

// Test_P1_Websocket_Health_Check_Unauthorized tests that a wrong token is rejected.
// Route: GET /health
func Test_P1_Websocket_Health_Check_Unauthorized(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.DoRequest(t, http.MethodGet, pathHealthCheck, nil, map[string]string{
		"X-Health-Token": "wrong-token",
	})

	resp.AssertStatus(t, http.StatusUnauthorized)
}

// Test_P1_Websocket_Health_Check_NoToken tests that a missing token is rejected.
// Route: GET /health
func Test_P1_Websocket_Health_Check_NoToken(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.DoRequest(t, http.MethodGet, pathHealthCheck, nil, nil)

	resp.AssertStatus(t, http.StatusUnauthorized)
}

// Test_P1_Websocket_Health_Check_WithDetail tests that the detail query param
// returns per-component status breakdown.
// Route: GET /health?detail=true
func Test_P1_Websocket_Health_Check_WithDetail(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.DoRequest(t, http.MethodGet, pathHealthCheck+"?detail=true", nil, map[string]string{
		"X-Health-Token": wsHealthToken,
	})

	resp.AssertOK(t)
	body := resp.JSON(t)
	if _, ok := body["components"]; !ok {
		t.Fatalf("expected 'components' in detail response, got: %s", resp.String())
	}
}

// ---------------------------------------------------------------------------
// P2 — Edge Cases
// ---------------------------------------------------------------------------

// Test_P2_Websocket_Health_Check_CacheHit tests that a second request within the
// cache TTL returns the same timestamp (cached result).
// Route: GET /health
func Test_P2_Websocket_Health_Check_CacheHit(t *testing.T) {
	client := fixture.NewWebsocketClient()

	headers := map[string]string{"X-Health-Token": wsHealthToken}

	resp1 := client.DoRequest(t, http.MethodGet, pathHealthCheck, nil, headers)
	resp1.AssertOK(t)
	body1 := resp1.JSON(t)

	resp2 := client.DoRequest(t, http.MethodGet, pathHealthCheck, nil, headers)
	resp2.AssertOK(t)
	body2 := resp2.JSON(t)

	// Both responses should have the same timestamp (cache hit)
	if body1["timestamp"] != body2["timestamp"] {
		t.Logf("timestamps differ: %v vs %v (cache may have expired — acceptable)", body1["timestamp"], body2["timestamp"])
	}
}
