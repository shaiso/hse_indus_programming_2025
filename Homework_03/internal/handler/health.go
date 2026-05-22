package handler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/shaiso/autodoc-api/internal/config"
	"github.com/shaiso/autodoc-api/internal/llm"
	"github.com/shaiso/autodoc-api/internal/version"
)

type llmStatus struct {
	Reachable bool   `json:"reachable"`
	Model     string `json:"model"`
	Error     string `json:"error,omitempty"`
}

type healthResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Commit  string    `json:"commit"`
	LLM     llmStatus `json:"llm"`
}

type HealthHandler struct {
	llm    llm.Client
	cfg    *config.Config
	mu     sync.Mutex
	cached llmStatus
	stamp  time.Time
	ttl    time.Duration
}

func NewHealthHandler(l llm.Client, cfg *config.Config) *HealthHandler {
	return &HealthHandler{llm: l, cfg: cfg, ttl: 30 * time.Second}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := h.llmCheck(r.Context())
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Version: version.Version,
		Commit:  version.Commit,
		LLM:     st,
	})
}

func (h *HealthHandler) llmCheck(ctx context.Context) llmStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.stamp.IsZero() && time.Since(h.stamp) < h.ttl {
		return h.cached
	}
	st := llmStatus{Model: h.llm.PrimaryModel()}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := h.llm.Ping(pingCtx); err != nil {
		st.Reachable = false
		st.Error = err.Error()
	} else {
		st.Reachable = true
	}
	h.cached = st
	h.stamp = time.Now()
	return st
}
