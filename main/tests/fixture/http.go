package fixture

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// HTTPClient wraps http.Client with convenience methods for testing.
type HTTPClient struct {
	client  *http.Client
	baseURL string
	token   string
}

// NewHTTPClient creates a new HTTP client for testing.
// It reads SERVICE_URL from environment or uses the provided baseURL.
func NewHTTPClient(baseURL ...string) *HTTPClient {
	url := ""
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	} else {
		url = getEnv("SERVICE_URL", "http://localhost:8080")
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: url,
	}
}

// WithToken sets the authorization token for subsequent requests.
func (c *HTTPClient) WithToken(token string) *HTTPClient {
	c.token = token
	return c
}

// Get performs a GET request to the given path.
func (c *HTTPClient) Get(tb testing.TB, path string, headers ...map[string]string) *HTTPResponse {
	tb.Helper()

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		tb.Fatalf("failed to create GET request: %v", err)
	}

	return c.doRequest(tb, req, headers...)
}

// Post performs a POST request to the given path with a JSON body.
func (c *HTTPClient) Post(tb testing.TB, path string, body any, headers ...map[string]string) *HTTPResponse {
	tb.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			tb.Fatalf("failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		tb.Fatalf("failed to create POST request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doRequest(tb, req, headers...)
}

// Delete performs a DELETE request to the given path with an optional JSON body.
func (c *HTTPClient) Delete(tb testing.TB, path string, body any, headers ...map[string]string) *HTTPResponse {
	tb.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			tb.Fatalf("failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(http.MethodDelete, c.baseURL+path, bodyReader)
	if err != nil {
		tb.Fatalf("failed to create DELETE request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doRequest(tb, req, headers...)
}

// doRequest executes the HTTP request and returns the response.
func (c *HTTPClient) doRequest(tb testing.TB, req *http.Request, headers ...map[string]string) *HTTPResponse {
	tb.Helper()

	// Set authorization header if token is set
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	// Set additional headers
	for _, h := range headers {
		for key, value := range h {
			req.Header.Set(key, value)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		tb.Fatalf("failed to execute request: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		tb.Fatalf("failed to read response body: %v", err)
	}
	resp.Body.Close()

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}
}

// HTTPResponse wraps an HTTP response for testing.
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// JSON unmarshals the response body into the given value.
func (r *HTTPResponse) JSON(tb testing.TB, v any) {
	tb.Helper()

	if err := json.Unmarshal(r.Body, v); err != nil {
		tb.Fatalf("failed to unmarshal response body: %v", err)
	}
}

// String returns the response body as a string.
func (r *HTTPResponse) String() string {
	return string(r.Body)
}

// AssertStatus asserts that the response has the expected status code.
func (r *HTTPResponse) AssertStatus(tb testing.TB, expected int) *HTTPResponse {
	tb.Helper()

	if r.StatusCode != expected {
		tb.Fatalf("expected status %d but got %d: %s", expected, r.StatusCode, r.String())
	}
	return r
}

// AssertOK asserts that the response status is 200 OK.
func (r *HTTPResponse) AssertOK(tb testing.TB) *HTTPResponse {
	tb.Helper()
	return r.AssertStatus(tb, http.StatusOK)
}

// AssertCreated asserts that the response status is 201 Created.
func (r *HTTPResponse) AssertCreated(tb testing.TB) *HTTPResponse {
	tb.Helper()
	return r.AssertStatus(tb, http.StatusCreated)
}

// AssertBadRequest asserts that the response status is 400 Bad Request.
func (r *HTTPResponse) AssertBadRequest(tb testing.TB) *HTTPResponse {
	tb.Helper()
	return r.AssertStatus(tb, http.StatusBadRequest)
}

// AssertUnauthorized asserts that the response status is 401 Unauthorized.
func (r *HTTPResponse) AssertUnauthorized(tb testing.TB) *HTTPResponse {
	tb.Helper()
	return r.AssertStatus(tb, http.StatusUnauthorized)
}

// AssertNotFound asserts that the response status is 404 Not Found.
func (r *HTTPResponse) AssertNotFound(tb testing.TB) *HTTPResponse {
	tb.Helper()
	return r.AssertStatus(tb, http.StatusNotFound)
}

// AssertInternalError asserts that the response status is 500 Internal Server Error.
func (r *HTTPResponse) AssertInternalError(tb testing.TB) *HTTPResponse {
	tb.Helper()
	return r.AssertStatus(tb, http.StatusInternalServerError)
}

// APIResponse represents the standard API response structure.
type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// ParseAPIResponse parses the response as a standard API response.
func (r *HTTPResponse) ParseAPIResponse(tb testing.TB) *APIResponse {
	tb.Helper()

	var resp APIResponse
	r.JSON(tb, &resp)
	return &resp
}

// AssertSuccess asserts that the API response indicates success (code == 0).
func (r *HTTPResponse) AssertSuccess(tb testing.TB) *HTTPResponse {
	tb.Helper()

	apiResp := r.ParseAPIResponse(tb)
	if apiResp.Code != 0 {
		tb.Fatalf("expected success code 0 but got %d: %s", apiResp.Code, apiResp.Message)
	}
	return r
}

// AssertErrorCode asserts that the API response has the expected error code.
func (r *HTTPResponse) AssertErrorCode(tb testing.TB, expectedCode int) *HTTPResponse {
	tb.Helper()

	apiResp := r.ParseAPIResponse(tb)
	if apiResp.Code != expectedCode {
		tb.Fatalf("expected error code %d but got %d: %s", expectedCode, apiResp.Code, apiResp.Message)
	}
	return r
}
