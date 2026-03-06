package ai_agent

import (
	"os"
	"strconv"
)

// Config holds AI agent configuration, loaded from environment variables.
type Config struct {
	// LLM settings
	LLMProvider    string // "openai" or "anthropic"
	LLMModel       string // e.g. "gpt-4o-mini"
	LLMAPIKey      string
	LLMBaseURL     string // OpenAI-compatible base URL
	LLMTemperature float64
	LLMMaxTokens   int

	// Agent behavior
	ForecastDays         int
	SafetyStockThreshold float64 // multiplier for safety buffer, e.g. 1.5
}

// DefaultConfig returns configuration with sensible defaults, overridden by env vars.
func DefaultConfig() Config {
	// Priority: AI_AGENT_LLM_* > LLM_* > default
	cfg := Config{
		LLMProvider:          getEnvWithFallbackKeys("openai", "AI_AGENT_LLM_PROVIDER", "LLM_PROVIDER"),
		LLMModel:             getEnvWithFallbackKeys("gpt-4o-mini", "AI_AGENT_LLM_MODEL", "LLM_MODEL"),
		LLMAPIKey:            getEnvWithFallbackKeys("", "AI_AGENT_LLM_API_KEY", "LLM_API_KEY"),
		LLMBaseURL:           getEnvWithFallbackKeys("https://api.openai.com/v1", "AI_AGENT_LLM_BASE_URL", "LLM_BASE_URL"),
		LLMTemperature:       getEnvFloat("AI_AGENT_LLM_TEMPERATURE", 0.2),
		LLMMaxTokens:         getEnvInt("AI_AGENT_LLM_MAX_TOKENS", 4096),
		ForecastDays:         getEnvInt("AI_AGENT_FORECAST_DAYS", 3),
		SafetyStockThreshold: getEnvFloat("AI_AGENT_SAFETY_THRESHOLD", 1.5),
	}
	// Fallback for temperature/max_tokens from LLM_* vars
	if cfg.LLMTemperature == 0.2 {
		if v := os.Getenv("LLM_TEMPERATURE"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				cfg.LLMTemperature = f
			}
		}
	}
	if cfg.LLMMaxTokens == 4096 {
		if v := os.Getenv("LLM_MAX_TOKENS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				cfg.LLMMaxTokens = n
			}
		}
	}
	return cfg
}

// getEnvWithFallbackKeys tries multiple env keys in order, returns first non-empty value or fallback.
func getEnvWithFallbackKeys(fallback string, keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
