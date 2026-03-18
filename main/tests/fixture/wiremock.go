package fixture

import (
	"sync"
	"testing"

	"github.com/hitosea/wiremock/client"
)

var (
	wmOnce   sync.Once
	wmClient *client.Client
)

// SetupWireMock returns a WireMock client with default TTPOS stubs registered.
// Default stubs are registered exactly once per process (safe for parallel packages).
// Named stubs are idempotent — duplicate registrations from concurrent processes are harmless.
//
// Tests that need custom stubs should register them at a lower priority number
// (higher priority) so they take precedence over the defaults (priority=100).
func SetupWireMock(tb testing.TB) *client.Client {
	tb.Helper()

	wmOnce.Do(func() {
		wmClient = client.NewClient(getEnv("WIREMOCK_URL", "http://wiremock:8080"))
		client.RegisterTTPOSDefaults(wmClient, tb)
		registerExtraDefaults(wmClient, tb)
	})

	return wmClient
}

// registerExtraDefaults adds stubs for gRPC methods not covered by the library defaults.
func registerExtraDefaults(c *client.Client, tb testing.TB) {
	tb.Helper()

	extras := []*client.StubMapping{
		// websocket.WebSocket — PushMessage (called by shift close to notify clients)
		client.NewGrpcStub("websocket.WebSocket", "PushMessage").
			WithName("default-push-message").
			WithPriority(100).
			WillReturnResponse(map[string]any{
				"success": true,
			}).
			Build(),
		// websocket.WebSocket — CheckDeviceOnline
		client.NewGrpcStub("websocket.WebSocket", "CheckDeviceOnline").
			WithName("default-check-device-online").
			WithPriority(100).
			WillReturnResponse(map[string]any{
				"online": false,
			}).
			Build(),
		// websocket.WebSocket — CloseConnection
		client.NewGrpcStub("websocket.WebSocket", "CloseConnection").
			WithName("default-close-connection").
			WithPriority(100).
			WillReturnResponse(map[string]any{
				"success": true,
			}).
			Build(),
	}
	client.RegisterDefaults(c, tb, extras)
}
