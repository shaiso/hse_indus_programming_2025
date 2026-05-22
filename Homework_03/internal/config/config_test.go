package config

import (
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-fake")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Port)
	}
	if cfg.LLMModel != "gpt-4.1-mini" {
		t.Errorf("model = %q", cfg.LLMModel)
	}
	if cfg.MaxFiles != 100 {
		t.Errorf("max files = %d", cfg.MaxFiles)
	}
}

func TestLoad_RequiresAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("USE_MOCK_LLM", "false")
	if _, err := Load(); err == nil {
		t.Fatal("expected error without OPENAI_API_KEY")
	}
}

func TestLoad_MockSkipsAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("USE_MOCK_LLM", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.UseMockLLM {
		t.Error("UseMockLLM should be true")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-fake")
	t.Setenv("LOG_LEVEL", "verbose")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-fake")
	t.Setenv("PORT", "999999")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid port")
	}
}
