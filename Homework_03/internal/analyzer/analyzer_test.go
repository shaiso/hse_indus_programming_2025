package analyzer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/shaiso/autodoc-api/internal/llm"
	"github.com/shaiso/autodoc-api/internal/parser"
	"github.com/shaiso/autodoc-api/internal/spec"
)

const goodSpec = `{
  "openapi": "3.0.3",
  "info": {"title": "T", "version": "1.0"},
  "paths": {
    "/ping": {"get": {"responses": {"200": {"description": "ok"}}}}
  }
}`

const badSpec = `{"openapi": "3.0.3", "info": {"title": "T"}}`

func newTestService(mock *llm.MockClient) *Service {
	return New(parser.New(), mock, spec.New(), nil)
}

func TestAnalyze_HappyPath(t *testing.T) {
	mock := llm.NewMock()
	mock.SetResponse(goodSpec)
	svc := newTestService(mock)

	res, err := svc.Analyze(context.Background(), Request{
		Files: []parser.File{{Name: "main.go", Content: `import "github.com/gin-gonic/gin"`}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.EndpointsCount != 1 {
		t.Errorf("endpoints = %d, want 1", res.EndpointsCount)
	}
	if res.Language != "go" {
		t.Errorf("language = %q, want go", res.Language)
	}
	if res.Framework != "gin" {
		t.Errorf("framework = %q, want gin", res.Framework)
	}
	if res.Model != "mock" {
		t.Errorf("model = %q, want mock", res.Model)
	}
}

func TestAnalyze_LLMError(t *testing.T) {
	mock := llm.NewMock()
	mock.SetError(llm.ErrUpstream)
	svc := newTestService(mock)

	_, err := svc.Analyze(context.Background(), Request{
		Files: []parser.File{{Name: "main.go", Content: "package main"}},
	})
	if err == nil {
		t.Fatal("expected llm error")
	}
	if !errors.Is(err, ErrLLM) {
		t.Errorf("error = %v, should wrap ErrLLM", err)
	}
}

func TestAnalyze_SpecRetry(t *testing.T) {
	mock := &spyMock{responses: []string{badSpec, goodSpec}}
	svc := New(parser.New(), mock, spec.New(), nil)

	res, err := svc.Analyze(context.Background(), Request{
		Files: []parser.File{{Name: "main.go", Content: "package main"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.RetriedSpec {
		t.Errorf("expected RetriedSpec=true")
	}
	if mock.calls != 2 {
		t.Errorf("expected 2 LLM calls, got %d", mock.calls)
	}
}

func TestAnalyze_EmptyInput(t *testing.T) {
	svc := newTestService(llm.NewMock())
	_, err := svc.Analyze(context.Background(), Request{Files: nil})
	if err == nil {
		t.Fatal("expected empty input error")
	}
	if !errors.Is(err, ErrEmpty) {
		t.Errorf("error = %v, want ErrEmpty", err)
	}
}

func TestAnalyze_TooLarge(t *testing.T) {
	svc := newTestService(llm.NewMock())
	_, err := svc.Analyze(context.Background(), Request{
		Files:          []parser.File{{Name: "a.go", Content: strings.Repeat("a", 1000)}},
		MaxRequestSize: 100,
	})
	if err == nil {
		t.Fatal("expected size error")
	}
	if !errors.Is(err, ErrTooLarge) {
		t.Errorf("error = %v, want ErrTooLarge", err)
	}
}

// spyMock returns different responses on subsequent calls.
type spyMock struct {
	responses []string
	calls     int
}

func (s *spyMock) Generate(_ context.Context, _ llm.Request) (llm.Response, error) {
	idx := s.calls
	if idx >= len(s.responses) {
		idx = len(s.responses) - 1
	}
	s.calls++
	return llm.Response{Content: s.responses[idx], Model: "spy"}, nil
}

func (s *spyMock) Ping(_ context.Context) error { return nil }

func (s *spyMock) PrimaryModel() string { return "spy" }
