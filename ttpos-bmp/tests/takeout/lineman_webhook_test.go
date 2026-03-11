//go:build integration

package takeout_test

import (
	"net/http"
	"testing"
	"time"

	"ttpos-bmp/tests/fixture"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	pathLinemanPlaceOrder  = "/api/v1/lmwn/partners/test-partner/stores/12345/orders"
	pathLinemanTriggerSync = "/api/v1/lmwn/partners/test-partner/stores/12345/menus/trigger-sync"
	pathLinemanOAuthToken  = "/api/v1/lmwn/oauth2/token"

	// linemanJWTSecret must match tests/config/takeout.yaml → app.provider.lineman.platform.secretKey
	linemanJWTSecret = "test-secret-key"
)

// ---------------------------------------------------------------------------
// P0 — Critical Path
// ---------------------------------------------------------------------------

// Test_P0_Takeout_LinemanWebhook_PlaceOrder_Unauthorized tests that a Lineman
// order request without an Authorization header is rejected by JWT middleware.
// Route: POST /api/v1/lmwn/partners/:partnerId/stores/:storeId/orders
func Test_P0_Takeout_LinemanWebhook_PlaceOrder_Unauthorized(t *testing.T) {
	client := fixture.NewTakeoutClient()

	resp := client.DoRequest(t, http.MethodPost, pathLinemanPlaceOrder, map[string]any{}, nil)

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
	resp.AssertBodyContains(t, "Authorization header is required")
}

// Test_P0_Takeout_LinemanWebhook_TriggerSyncMenu_Unauthorized tests that a
// menu sync trigger without auth is rejected by JWT middleware.
// Route: POST /api/v1/lmwn/partners/:partnerId/stores/:storeId/menus/trigger-sync
func Test_P0_Takeout_LinemanWebhook_TriggerSyncMenu_Unauthorized(t *testing.T) {
	client := fixture.NewTakeoutClient()

	resp := client.DoRequest(t, http.MethodPost, pathLinemanTriggerSync, map[string]any{}, nil)

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
	resp.AssertBodyContains(t, "Authorization header is required")
}

// Test_P0_Takeout_LinemanWebhook_OAuthToken_NoAuthRequired tests that the OAuth
// token endpoint is exempt from JWT auth (no Authorization header needed).
// The endpoint returns 400 validation error for missing grant_type, not 401 auth error.
// Route: POST /api/v1/lmwn/oauth2/token
func Test_P0_Takeout_LinemanWebhook_OAuthToken_NoAuthRequired(t *testing.T) {
	client := fixture.NewTakeoutClient()

	// OAuth endpoint is exempt from JWT auth — returns validation error (400), not auth error (401)
	resp := client.DoRequest(t, http.MethodPost, pathLinemanOAuthToken, map[string]any{}, nil)

	// Expect 400 (validation error: missing grant_type), not 401 (auth error)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 (validation error), got %d: %s", resp.StatusCode, resp.String())
	}
	body := resp.JSON(t)
	if errField, ok := body["error"]; ok && errField == "unauthorized" {
		if desc, _ := body["error_description"].(string); desc == "Authorization header is required" {
			t.Fatalf("OAuth token endpoint should not require auth, got: %s", resp.String())
		}
	}
	// Should have validation error message about grant_type
	if msg, _ := body["message"].(string); msg != "" && msg != "授权类型不能为空" {
		t.Logf("got validation message: %s", msg)
	}
}

// ---------------------------------------------------------------------------
// P1 — Important Validation
// ---------------------------------------------------------------------------

// Test_P1_Takeout_LinemanWebhook_PlaceOrder_TokenExpired tests that an expired
// Lineman JWT is rejected by the middleware.
// Route: POST /api/v1/lmwn/partners/:partnerId/stores/:storeId/orders
func Test_P1_Takeout_LinemanWebhook_PlaceOrder_TokenExpired(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT(linemanJWTSecret, -1*time.Minute) // already expired
	resp := client.DoRequest(t, http.MethodPost, pathLinemanPlaceOrder, map[string]any{}, map[string]string{
		"Authorization": "Bearer " + token,
	})

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
}

// Test_P1_Takeout_LinemanWebhook_PlaceOrder_TokenInvalid tests that a JWT signed
// with the wrong secret is rejected.
// Route: POST /api/v1/lmwn/partners/:partnerId/stores/:storeId/orders
func Test_P1_Takeout_LinemanWebhook_PlaceOrder_TokenInvalid(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT("wrong-secret-key", 5*time.Minute)
	resp := client.DoRequest(t, http.MethodPost, pathLinemanPlaceOrder, map[string]any{}, map[string]string{
		"Authorization": "Bearer " + token,
	})

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
}

// Test_P1_Takeout_LinemanWebhook_TriggerSyncMenu_ValidToken tests that a valid
// JWT token passes auth on the menu sync endpoint.
// Route: POST /api/v1/lmwn/partners/:partnerId/stores/:storeId/menus/trigger-sync
func Test_P1_Takeout_LinemanWebhook_TriggerSyncMenu_ValidToken(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT(linemanJWTSecret, 5*time.Minute)
	resp := client.DoRequest(t, http.MethodPost, pathLinemanTriggerSync, map[string]any{}, map[string]string{
		"Authorization": "Bearer " + token,
	})

	resp.AssertOK(t)
	body := resp.JSON(t)
	if errField, ok := body["error"]; ok && errField == "unauthorized" {
		t.Fatalf("expected auth to pass, got unauthorized: %s", resp.String())
	}
}

// ---------------------------------------------------------------------------
// P2 — Edge Cases
// ---------------------------------------------------------------------------

// Test_P2_Takeout_LinemanWebhook_PlaceOrder_EmptyBody tests that a valid JWT
// with an empty body reaches validation (not auth rejection).
// Route: POST /api/v1/lmwn/partners/:partnerId/stores/:storeId/orders
func Test_P2_Takeout_LinemanWebhook_PlaceOrder_EmptyBody(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT(linemanJWTSecret, 5*time.Minute)
	resp := client.DoRequest(t, http.MethodPost, pathLinemanPlaceOrder, nil, map[string]string{
		"Authorization": "Bearer " + token,
	})

	// Expect 400 (validation error: missing order_id), not 401 (auth error)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 (validation error), got %d: %s", resp.StatusCode, resp.String())
	}
	body := resp.JSON(t)
	if errField, ok := body["error"]; ok && errField == "unauthorized" {
		t.Fatalf("expected validation error, not auth rejection: %s", resp.String())
	}
	// Should have validation error message about order_id
	if msg, _ := body["message"].(string); msg != "" && msg != "订单ID不能为空" {
		t.Logf("got validation message: %s", msg)
	}
}
