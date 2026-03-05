package ai_agent

import (
	"os"
	"testing"
)

func TestDefaultConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.LLMProvider != "openai" {
		t.Errorf("expected provider=openai, got %s", cfg.LLMProvider)
	}
	if cfg.LLMModel != "gpt-4o-mini" {
		t.Errorf("expected model=gpt-4o-mini, got %s", cfg.LLMModel)
	}
	if cfg.ForecastDays != 3 {
		t.Errorf("expected forecast_days=3, got %d", cfg.ForecastDays)
	}
	if cfg.SafetyStockThreshold != 1.5 {
		t.Errorf("expected safety_threshold=1.5, got %.1f", cfg.SafetyStockThreshold)
	}
	if cfg.LLMTemperature != 0.2 {
		t.Errorf("expected temperature=0.2, got %.1f", cfg.LLMTemperature)
	}
	if cfg.LLMMaxTokens != 4096 {
		t.Errorf("expected max_tokens=4096, got %d", cfg.LLMMaxTokens)
	}
}

func TestDefaultConfig_EnvOverrides(t *testing.T) {
	os.Setenv("AI_AGENT_LLM_MODEL", "gpt-4o")
	os.Setenv("AI_AGENT_FORECAST_DAYS", "7")
	os.Setenv("AI_AGENT_SAFETY_THRESHOLD", "2.0")
	defer func() {
		os.Unsetenv("AI_AGENT_LLM_MODEL")
		os.Unsetenv("AI_AGENT_FORECAST_DAYS")
		os.Unsetenv("AI_AGENT_SAFETY_THRESHOLD")
	}()

	cfg := DefaultConfig()

	if cfg.LLMModel != "gpt-4o" {
		t.Errorf("expected model=gpt-4o, got %s", cfg.LLMModel)
	}
	if cfg.ForecastDays != 7 {
		t.Errorf("expected forecast_days=7, got %d", cfg.ForecastDays)
	}
	if cfg.SafetyStockThreshold != 2.0 {
		t.Errorf("expected safety_threshold=2.0, got %.1f", cfg.SafetyStockThreshold)
	}
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	os.Setenv("AI_AGENT_TEST_INT", "not_a_number")
	defer os.Unsetenv("AI_AGENT_TEST_INT")

	result := getEnvInt("AI_AGENT_TEST_INT", 42)
	if result != 42 {
		t.Errorf("expected fallback 42, got %d", result)
	}
}

func TestGetEnvFloat_InvalidValue(t *testing.T) {
	os.Setenv("AI_AGENT_TEST_FLOAT", "not_a_float")
	defer os.Unsetenv("AI_AGENT_TEST_FLOAT")

	result := getEnvFloat("AI_AGENT_TEST_FLOAT", 3.14)
	if result != 3.14 {
		t.Errorf("expected fallback 3.14, got %.2f", result)
	}
}
