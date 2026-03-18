//go:build integration

package takeout_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"ttpos-bmp/tests/fixture"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	pathGrabSubmitOrder = "/api/v1/grab/partner/orders"

	// grabJWTSecret must match tests/config/takeout.yaml → app.provider.grab.platform.secretKey
	grabJWTSecret = "test-secret-key"
)

// ---------------------------------------------------------------------------
// P0 — Critical Path
// ---------------------------------------------------------------------------

// Test_P0_Takeout_GrabWebhook_SubmitOrder_Unauthorized tests that requests
// without an Authorization header are rejected by the Grab JWT middleware.
// Route: POST /api/v1/grab/partner/orders
func Test_P0_Takeout_GrabWebhook_SubmitOrder_Unauthorized(t *testing.T) {
	client := fixture.NewTakeoutClient()

	resp := client.DoRequest(t, http.MethodPost, pathGrabSubmitOrder, map[string]any{}, nil)

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
	resp.AssertBodyContains(t, "Authorization header is required")
}

// Test_P0_Takeout_GrabWebhook_SubmitOrder_HappyPath tests that a request
// with a valid Grab JWT token passes auth and reaches business logic.
// Route: POST /api/v1/grab/partner/orders
func Test_P0_Takeout_GrabWebhook_SubmitOrder_HappyPath(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT(grabJWTSecret, 5*time.Minute)
	resp := client.DoRequest(t, http.MethodPost, pathGrabSubmitOrder, map[string]any{}, map[string]string{
		"Authorization": "Bearer " + token,
	})

	// Auth passes — may get business logic error (no order data), but not auth error
	resp.AssertOK(t)
	body := resp.JSON(t)
	if errField, ok := body["error"]; ok && errField == "unauthorized" {
		t.Fatalf("expected auth to pass, got unauthorized: %s", resp.String())
	}
}

// ---------------------------------------------------------------------------
// P1 — Important Validation
// ---------------------------------------------------------------------------

// Test_P1_Takeout_GrabWebhook_SubmitOrder_TokenExpired tests that an expired
// Grab JWT is rejected by the middleware.
// Route: POST /api/v1/grab/partner/orders
func Test_P1_Takeout_GrabWebhook_SubmitOrder_TokenExpired(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT(grabJWTSecret, -1*time.Minute) // already expired
	resp := client.DoRequest(t, http.MethodPost, pathGrabSubmitOrder, map[string]any{}, map[string]string{
		"Authorization": "Bearer " + token,
	})

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
}

// Test_P1_Takeout_GrabWebhook_SubmitOrder_TokenInvalid tests that a JWT signed
// with the wrong secret key is rejected.
// Route: POST /api/v1/grab/partner/orders
func Test_P1_Takeout_GrabWebhook_SubmitOrder_TokenInvalid(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT("wrong-secret-key", 5*time.Minute)
	resp := client.DoRequest(t, http.MethodPost, pathGrabSubmitOrder, map[string]any{}, map[string]string{
		"Authorization": "Bearer " + token,
	})

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
}

// Test_P1_Takeout_GrabWebhook_SubmitOrder_MalformedHeader tests that a malformed
// Authorization header (not "Bearer <token>" format) is rejected.
// Route: POST /api/v1/grab/partner/orders
func Test_P1_Takeout_GrabWebhook_SubmitOrder_MalformedHeader(t *testing.T) {
	client := fixture.NewTakeoutClient()

	resp := client.DoRequest(t, http.MethodPost, pathGrabSubmitOrder, map[string]any{}, map[string]string{
		"Authorization": "not-bearer-format",
	})

	resp.AssertOK(t)
	resp.AssertBodyContains(t, "unauthorized")
}

// ---------------------------------------------------------------------------
// P2 — Edge Cases
// ---------------------------------------------------------------------------

// Test_P2_Takeout_GrabWebhook_SubmitOrder_EmptyBody tests that a valid JWT with
// an empty body returns a validation error (not an auth error).
// Route: POST /api/v1/grab/partner/orders
func Test_P2_Takeout_GrabWebhook_SubmitOrder_EmptyBody(t *testing.T) {
	client := fixture.NewTakeoutClient()

	token := generateGrabJWT(grabJWTSecret, 5*time.Minute)
	resp := client.DoRequest(t, http.MethodPost, pathGrabSubmitOrder, nil, map[string]string{
		"Authorization": "Bearer " + token,
	})

	resp.AssertOK(t)
	// Should reach business logic, not be rejected by auth
	body := resp.JSON(t)
	if errField, ok := body["error"]; ok && errField == "unauthorized" {
		t.Fatalf("expected validation error, not auth rejection: %s", resp.String())
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// generateGrabJWT creates a minimal HS256 JWT for Grab webhook auth testing.
// Uses stdlib only (no external JWT library) to keep tests/go.mod dependency-free.
func generateGrabJWT(secret string, expiry time.Duration) string {
	header := base64.RawURLEncoding.EncodeToString(mustJSON(map[string]any{
		"alg": "HS256",
		"typ": "JWT",
	}))
	payload := base64.RawURLEncoding.EncodeToString(mustJSON(map[string]any{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(expiry).Unix(),
		"sub": "grab-partner-test",
	}))
	sigInput := fmt.Sprintf("%s.%s", header, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return strings.Join([]string{header, payload, sig}, ".")
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
