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
	pathWsPush = "/ws/push"

	// wsAPIKey must match tests/config/websocket.yaml → jwt.secret
	wsAPIKey = "test-ws-api-key"
)

// ---------------------------------------------------------------------------
// P0 — Critical Path
// ---------------------------------------------------------------------------

// Test_P0_Websocket_PushMessage_Unauthorized tests that a push request without
// an API key is rejected with code=500.
// Route: POST /ws/push
func Test_P0_Websocket_PushMessage_Unauthorized(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.Post(t, pathWsPush, map[string]any{
		"topic":   "test-topic",
		"message": "hello",
	})

	resp.AssertOK(t)
	body := resp.JSON(t)
	code, _ := body["code"].(float64)
	if code != 500 {
		t.Fatalf("expected code=500 for missing API key, got code=%v: %s", code, resp.String())
	}
	// Verify it's actually an auth-related error
	msg, _ := body["message"].(string)
	if msg != "KEY 不正确" {
		t.Fatalf("expected auth error message, got: %s", msg)
	}
}

// ---------------------------------------------------------------------------
// P1 — Important Validation
// ---------------------------------------------------------------------------

// Test_P1_Websocket_PushMessage_InvalidKey tests that a wrong API key is rejected.
// Route: POST /ws/push
func Test_P1_Websocket_PushMessage_InvalidKey(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.DoRequest(t, http.MethodPost, pathWsPush, map[string]any{
		"topic":   "test-topic",
		"message": "hello",
	}, map[string]string{
		"X-API-KEY": "wrong-api-key",
	})

	resp.AssertOK(t)
	body := resp.JSON(t)
	code, _ := body["code"].(float64)
	if code != 500 {
		t.Fatalf("expected code=500 for invalid API key, got code=%v: %s", code, resp.String())
	}
	// Verify it's actually an auth-related error
	msg, _ := body["message"].(string)
	if msg != "KEY 不正确" {
		t.Fatalf("expected auth error message, got: %s", msg)
	}
}

// Test_P1_Websocket_PushMessage_ValidKey tests that a valid API key passes auth
// and reaches business logic.
// Route: POST /ws/push
func Test_P1_Websocket_PushMessage_ValidKey(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.DoRequest(t, http.MethodPost, pathWsPush, map[string]any{
		"topic":   "test-topic",
		"message": "hello",
	}, map[string]string{
		"X-API-KEY": wsAPIKey,
	})

	resp.AssertOK(t)
	body := resp.JSON(t)
	// Auth passes — code should not be -1 (auth rejection)
	code, _ := body["code"].(float64)
	if code == -1 {
		if msg, _ := body["message"].(string); msg == "KEY 不正确" {
			t.Fatalf("expected auth to pass, got API key rejection: %s", resp.String())
		}
	}
}

// ---------------------------------------------------------------------------
// P2 — Edge Cases
// ---------------------------------------------------------------------------

// Test_P2_Websocket_PushMessage_EmptyBody tests that a valid key with an empty
// body reaches validation (not auth rejection).
// Route: POST /ws/push
func Test_P2_Websocket_PushMessage_EmptyBody(t *testing.T) {
	client := fixture.NewWebsocketClient()

	resp := client.DoRequest(t, http.MethodPost, pathWsPush, nil, map[string]string{
		"X-API-KEY": wsAPIKey,
	})

	resp.AssertOK(t)
	body := resp.JSON(t)
	// Should not be auth rejection
	if msg, _ := body["message"].(string); msg == "KEY 不正确" {
		t.Fatalf("expected validation error, not auth rejection: %s", resp.String())
	}
}
