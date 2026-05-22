package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port             int
	LogLevel         string
	OpenAIAPIKey     string
	OpenAIBaseURL    string
	LLMModel         string
	LLMFallbackModel string
	LLMTimeout       time.Duration
	MaxRequestBytes  int64
	MaxFiles         int
	RateLimitPerMin  int
	UseMockLLM       bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnvInt("PORT", 8080),
		LogLevel:         getEnvStr("LOG_LEVEL", "info"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:    os.Getenv("OPENAI_BASE_URL"),
		LLMModel:         getEnvStr("LLM_MODEL", "gpt-4.1-mini"),
		LLMFallbackModel: getEnvStr("LLM_FALLBACK_MODEL", "gpt-4.1"),
		LLMTimeout:       getEnvDuration("LLM_TIMEOUT", 60*time.Second),
		MaxRequestBytes:  int64(getEnvInt("MAX_REQUEST_BYTES", 2*1024*1024)),
		MaxFiles:         getEnvInt("MAX_FILES", 100),
		RateLimitPerMin:  getEnvInt("RATE_LIMIT_PER_MIN", 60),
		UseMockLLM:       getEnvBool("USE_MOCK_LLM", false),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid PORT: %d", c.Port)
	}
	if !c.UseMockLLM && c.OpenAIAPIKey == "" {
		return errors.New("OPENAI_API_KEY is required (or set USE_MOCK_LLM=true)")
	}
	if c.MaxRequestBytes <= 0 {
		return errors.New("MAX_REQUEST_BYTES must be positive")
	}
	if c.MaxFiles <= 0 {
		return errors.New("MAX_FILES must be positive")
	}
	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL: %s", c.LogLevel)
	}
	return nil
}

func getEnvStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
