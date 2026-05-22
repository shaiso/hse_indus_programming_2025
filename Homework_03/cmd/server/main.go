package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shaiso/autodoc-api/internal/analyzer"
	"github.com/shaiso/autodoc-api/internal/config"
	"github.com/shaiso/autodoc-api/internal/handler"
	"github.com/shaiso/autodoc-api/internal/llm"
	"github.com/shaiso/autodoc-api/internal/middleware"
	"github.com/shaiso/autodoc-api/internal/parser"
	"github.com/shaiso/autodoc-api/internal/spec"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	logger.Info("starting AutoDoc API",
		"port", cfg.Port,
		"model", cfg.LLMModel,
		"mock_llm", cfg.UseMockLLM,
	)

	llmClient := buildLLMClient(cfg, logger)
	parserSvc := parser.New()
	specBuilder := spec.New()
	analyzerSvc := analyzer.New(parserSvc, llmClient, specBuilder, logger)

	httpHandler := buildHandler(cfg, logger, analyzerSvc, llmClient)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           httpHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("shutdown complete")
	return nil
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}

func buildLLMClient(cfg *config.Config, logger *slog.Logger) llm.Client {
	if cfg.UseMockLLM {
		logger.Warn("LLM client is in MOCK mode")
		return llm.NewMock()
	}
	return llm.NewOpenAI(llm.OpenAIConfig{
		APIKey:        cfg.OpenAIAPIKey,
		BaseURL:       cfg.OpenAIBaseURL,
		Model:         cfg.LLMModel,
		FallbackModel: cfg.LLMFallbackModel,
		Timeout:       cfg.LLMTimeout,
		Logger:        logger,
	})
}

func buildHandler(cfg *config.Config, logger *slog.Logger, a *analyzer.Service, llmClient llm.Client) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/analyze", handler.NewAnalyzeHandler(a, cfg))
	mux.Handle("GET /api/v1/health", handler.NewHealthHandler(llmClient, cfg))

	return middleware.Chain(mux,
		middleware.Recovery(logger),
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.RateLimit(cfg.RateLimitPerMin),
		middleware.RequestSizeLimit(cfg.MaxRequestBytes),
	)
}
