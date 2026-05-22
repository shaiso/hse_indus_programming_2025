package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shaiso/autodoc-api/internal/analyzer"
	"github.com/shaiso/autodoc-api/internal/config"
	"github.com/shaiso/autodoc-api/internal/llm"
	"github.com/shaiso/autodoc-api/internal/parser"
	"github.com/shaiso/autodoc-api/internal/spec"
)

const validMockSpec = `{
  "openapi": "3.0.3",
  "info": {"title": "T", "version": "1.0"},
  "paths": {"/x": {"get": {"responses": {"200": {"description": "ok"}}}}}
}`

func newTestMux(mock *llm.MockClient) (http.Handler, *config.Config) {
	cfg := &config.Config{
		MaxFiles:        100,
		MaxRequestBytes: 2 * 1024 * 1024,
	}
	a := analyzer.New(parser.New(), mock, spec.New(), nil)
	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/analyze", NewAnalyzeHandler(a, cfg))
	mux.Handle("GET /api/v1/health", NewHealthHandler(mock, cfg))
	return mux, cfg
}

func doJSON(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestAnalyze_OK(t *testing.T) {
	mock := llm.NewMock()
	mock.SetResponse(validMockSpec)
	h, _ := newTestMux(mock)

	w := doJSON(h, "POST", "/api/v1/analyze", map[string]any{
		"files":  []map[string]string{{"name": "main.go", "content": "package main"}},
		"format": "yaml",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["format"] != "yaml" {
		t.Errorf("format = %v, want yaml", resp["format"])
	}
	if resp["endpoints_count"].(float64) != 1 {
		t.Errorf("endpoints_count = %v", resp["endpoints_count"])
	}
}

func TestAnalyze_EmptyFiles_422(t *testing.T) {
	mock := llm.NewMock()
	h, _ := newTestMux(mock)

	w := doJSON(h, "POST", "/api/v1/analyze", map[string]any{
		"files": []map[string]string{},
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestAnalyze_InvalidFormat_422(t *testing.T) {
	mock := llm.NewMock()
	h, _ := newTestMux(mock)
	w := doJSON(h, "POST", "/api/v1/analyze", map[string]any{
		"files":  []map[string]string{{"name": "a.go", "content": "x"}},
		"format": "xml",
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestAnalyze_UnsupportedLanguage_422(t *testing.T) {
	mock := llm.NewMock()
	h, _ := newTestMux(mock)
	w := doJSON(h, "POST", "/api/v1/analyze", map[string]any{
		"files":    []map[string]string{{"name": "a.go", "content": "x"}},
		"language": "rust",
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422, body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "unsupported_language" {
		t.Errorf("code = %v, want unsupported_language", resp["code"])
	}
}

func TestAnalyze_LLMError_502(t *testing.T) {
	mock := llm.NewMock()
	mock.SetError(errors.New("network down"))
	mock.SetError(llm.ErrUpstream)
	h, _ := newTestMux(mock)

	w := doJSON(h, "POST", "/api/v1/analyze", map[string]any{
		"files": []map[string]string{{"name": "a.go", "content": "x"}},
	})
	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502, body=%s", w.Code, w.Body.String())
	}
}

func TestAnalyze_TooManyFiles_413(t *testing.T) {
	mock := llm.NewMock()
	h, cfg := newTestMux(mock)
	cfg.MaxFiles = 2

	files := []map[string]string{
		{"name": "a.go", "content": "x"},
		{"name": "b.go", "content": "y"},
		{"name": "c.go", "content": "z"},
	}
	w := doJSON(h, "POST", "/api/v1/analyze", map[string]any{"files": files})
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", w.Code)
	}
}

func TestAnalyze_MalformedJSON_422(t *testing.T) {
	mock := llm.NewMock()
	h, _ := newTestMux(mock)
	req := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHealth_OK(t *testing.T) {
	mock := llm.NewMock()
	h, _ := newTestMux(mock)
	w := doJSON(h, "GET", "/api/v1/health", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %v", resp["status"])
	}
	llmField := resp["llm"].(map[string]any)
	if llmField["reachable"] != true {
		t.Errorf("llm.reachable = %v", llmField["reachable"])
	}
}

func TestHealth_LLMUnreachable(t *testing.T) {
	mock := llm.NewMock()
	mock.PingErr = errors.New("upstream down")
	h, _ := newTestMux(mock)
	w := doJSON(h, "GET", "/api/v1/health", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	llmField := resp["llm"].(map[string]any)
	if llmField["reachable"] != false {
		t.Errorf("expected reachable=false, got %v", llmField["reachable"])
	}
}
