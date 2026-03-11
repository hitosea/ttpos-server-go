package fixture

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// HTTPClient wraps http.Client with convenience methods for BMP service testing.
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewErpClient creates an HTTP client targeting the ERP service.
func NewErpClient() *HTTPClient {
	return newHTTPClient(getEnv("ERP_URL", "http://localhost:14021"))
}

// NewTakeoutClient creates an HTTP client targeting the Takeout service.
func NewTakeoutClient() *HTTPClient {
	return newHTTPClient(getEnv("TAKEOUT_URL", "http://localhost:14031"))
}

// NewMessageClient creates an HTTP client targeting the Message service.
func NewMessageClient() *HTTPClient {
	return newHTTPClient(getEnv("MESSAGE_URL", "http://localhost:14041"))
}

// NewWebsocketClient creates an HTTP client targeting the WebSocket service.
func NewWebsocketClient() *HTTPClient {
	return newHTTPClient(getEnv("WEBSOCKET_URL", "http://localhost:14051"))
}

func newHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

// Get performs a GET request to the given path.
func (c *HTTPClient) Get(tb testing.TB, path string) *HTTPResponse {
	tb.Helper()
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		tb.Fatalf("failed to create GET request: %v", err)
	}
	return c.do(tb, req)
}

// Post performs a POST request with a JSON body.
func (c *HTTPClient) Post(tb testing.TB, path string, body any) *HTTPResponse {
	tb.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			tb.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		tb.Fatalf("failed to create POST request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(tb, req)
}

// DoRequest performs a request with a given method, path, optional JSON body, and custom headers.
func (c *HTTPClient) DoRequest(tb testing.TB, method, path string, body any, headers map[string]string) *HTTPResponse {
	tb.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			tb.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		tb.Fatalf("failed to create %s request: %v", method, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.do(tb, req)
}

func (c *HTTPClient) do(tb testing.TB, req *http.Request) *HTTPResponse {
	tb.Helper()
	resp, err := c.client.Do(req)
	if err != nil {
		tb.Fatalf("request failed: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		tb.Fatalf("failed to read response body: %v", err)
	}
	return &HTTPResponse{StatusCode: resp.StatusCode, Body: body}
}

// HTTPResponse wraps an HTTP response for assertions.
type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

// String returns the response body as a string.
func (r *HTTPResponse) String() string {
	return string(r.Body)
}

// JSON parses the response body as a JSON object.
func (r *HTTPResponse) JSON(tb testing.TB) map[string]any {
	tb.Helper()
	var result map[string]any
	if err := json.Unmarshal(r.Body, &result); err != nil {
		tb.Fatalf("failed to parse response JSON: %v\nbody: %s", err, r.String())
	}
	return result
}

// AssertStatus asserts the expected HTTP status code.
func (r *HTTPResponse) AssertStatus(tb testing.TB, expected int) *HTTPResponse {
	tb.Helper()
	if r.StatusCode != expected {
		tb.Fatalf("expected status %d, got %d: %s", expected, r.StatusCode, r.String())
	}
	return r
}

// AssertOK asserts the response is 200 OK.
func (r *HTTPResponse) AssertOK(tb testing.TB) *HTTPResponse {
	tb.Helper()
	return r.AssertStatus(tb, http.StatusOK)
}

// AssertBodyContains asserts the response body contains the given substring.
func (r *HTTPResponse) AssertBodyContains(tb testing.TB, sub string) *HTTPResponse {
	tb.Helper()
	if !strings.Contains(r.String(), sub) {
		tb.Fatalf("expected body to contain %q, got: %s", sub, r.String())
	}
	return r
}
