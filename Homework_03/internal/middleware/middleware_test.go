package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestSizeLimit_ContentLength(t *testing.T) {
	h := RequestSizeLimit(10)(okHandler())

	body := strings.Repeat("a", 100)
	req := httptest.NewRequest("POST", "/x", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", w.Code)
	}
}

func TestRateLimit_Blocks(t *testing.T) {
	h := RateLimit(2)(okHandler())

	allowed, blocked := 0, 0
	for range 5 {
		req := httptest.NewRequest("GET", "/x", nil)
		req.RemoteAddr = "1.2.3.4:1000"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		switch w.Code {
		case http.StatusOK:
			allowed++
		case http.StatusTooManyRequests:
			blocked++
		}
	}
	if allowed > 2 {
		t.Errorf("allowed = %d, expected <= 2", allowed)
	}
	if blocked == 0 {
		t.Errorf("expected some blocked requests, got 0")
	}
}

func TestRequestID_Generated(t *testing.T) {
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, GetRequestID(r.Context()))
	}))

	req := httptest.NewRequest("GET", "/x", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Header().Get(RequestIDHeader) == "" {
		t.Error("expected X-Request-ID header to be set")
	}
	if w.Body.String() == "" {
		t.Error("expected body to contain generated request id")
	}
}

func TestRequestID_PassThrough(t *testing.T) {
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, GetRequestID(r.Context()))
	}))

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set(RequestIDHeader, "client-123")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Body.String() != "client-123" {
		t.Errorf("request id = %q, want client-123", w.Body.String())
	}
}
