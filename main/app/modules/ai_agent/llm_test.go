package ai_agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewLLMClient(t *testing.T) {
	cfg := Config{
		LLMBaseURL:     "https://api.example.com/v1",
		LLMAPIKey:      "test-key",
		LLMModel:       "gpt-4o-mini",
		LLMTemperature: 0.2,
		LLMMaxTokens:   4096,
	}
	client := NewLLMClient(cfg)

	if client.baseURL != cfg.LLMBaseURL {
		t.Errorf("expected baseURL=%s, got %s", cfg.LLMBaseURL, client.baseURL)
	}
	if client.apiKey != cfg.LLMAPIKey {
		t.Errorf("expected apiKey=%s, got %s", cfg.LLMAPIKey, client.apiKey)
	}
	if client.model != cfg.LLMModel {
		t.Errorf("expected model=%s, got %s", cfg.LLMModel, client.model)
	}
}

func TestLLMClient_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		// Verify request body
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "gpt-4o-mini" {
			t.Errorf("expected model=gpt-4o-mini, got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("expected system role, got %s", req.Messages[0].Role)
		}

		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "[]"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &LLMClient{
		baseURL:     server.URL,
		apiKey:      "test-key",
		model:       "gpt-4o-mini",
		temperature: 0.2,
		maxTokens:   4096,
		httpClient:  server.Client(),
	}

	result, err := client.Chat("You are a helper", "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "[]" {
		t.Errorf("expected [], got %s", result)
	}
}

func TestLLMClient_Chat_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := &LLMClient{
		baseURL:    server.URL,
		apiKey:     "test-key",
		model:      "gpt-4o-mini",
		httpClient: server.Client(),
	}

	_, err := client.Chat("system", "user")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestLLMClient_Chat_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{Choices: []chatChoice{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &LLMClient{
		baseURL:    server.URL,
		apiKey:     "test-key",
		model:      "gpt-4o-mini",
		httpClient: server.Client(),
	}

	_, err := client.Chat("system", "user")
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestLLMClient_Chat_ErrorField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": {"message": "rate limit exceeded"}, "choices": []}`))
	}))
	defer server.Close()

	client := &LLMClient{
		baseURL:    server.URL,
		apiKey:     "test-key",
		model:      "gpt-4o-mini",
		httpClient: server.Client(),
	}

	_, err := client.Chat("system", "user")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}
