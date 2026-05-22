package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/shaiso/autodoc-api/internal/analyzer"
	"github.com/shaiso/autodoc-api/internal/config"
	"github.com/shaiso/autodoc-api/internal/parser"
	"github.com/shaiso/autodoc-api/internal/spec"
)

type analyzeRequest struct {
	Files    []fileDTO `json:"files"`
	Language string    `json:"language"`
	Format   string    `json:"format"`
}

type fileDTO struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type tokensUsed struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type latencyBreakdown struct {
	TotalMS  int64 `json:"total_ms"`
	ParserMS int64 `json:"parser_ms"`
	LLMMS    int64 `json:"llm_ms"`
	SpecMS   int64 `json:"spec_ms"`
}

type analyzeResponse struct {
	Spec           string           `json:"spec"`
	Summary        string           `json:"summary"`
	EndpointsCount int              `json:"endpoints_count"`
	Format         string           `json:"format"`
	Language       string           `json:"language"`
	Framework      string           `json:"framework,omitempty"`
	Model          string           `json:"model"`
	UsedFallback   bool             `json:"used_fallback"`
	RetriedSpec    bool             `json:"retried_spec"`
	TokensUsed     tokensUsed       `json:"tokens_used"`
	Latency        latencyBreakdown `json:"latency"`
}

type AnalyzeHandler struct {
	analyzer *analyzer.Service
	cfg      *config.Config
}

func NewAnalyzeHandler(a *analyzer.Service, cfg *config.Config) *AnalyzeHandler {
	return &AnalyzeHandler{analyzer: a, cfg: cfg}
}

func (h *AnalyzeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req analyzeRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "input_too_large",
				"request body exceeds size limit", err.Error())
			return
		}
		writeError(w, http.StatusUnprocessableEntity, "invalid_format",
			"request body is malformed", err.Error())
		return
	}

	if len(req.Files) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "invalid_format",
			"files must contain at least one entry", "")
		return
	}
	for i, f := range req.Files {
		if strings.TrimSpace(f.Name) == "" {
			writeError(w, http.StatusUnprocessableEntity, "invalid_format",
				fmt.Sprintf("files[%d].name is required", i), "")
			return
		}
		if f.Content == "" {
			writeError(w, http.StatusUnprocessableEntity, "invalid_format",
				fmt.Sprintf("files[%d].content is required", i), "")
			return
		}
	}

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = spec.FormatYAML
	}
	if format != spec.FormatYAML && format != spec.FormatJSON {
		writeError(w, http.StatusUnprocessableEntity, "invalid_format",
			"format must be 'yaml' or 'json'", "")
		return
	}

	files := make([]parser.File, len(req.Files))
	for i, f := range req.Files {
		files[i] = parser.File{Name: f.Name, Content: f.Content}
	}

	result, err := h.analyzer.Analyze(r.Context(), analyzer.Request{
		Files:          files,
		Language:       req.Language,
		Format:         format,
		MaxFiles:       h.cfg.MaxFiles,
		MaxRequestSize: h.cfg.MaxRequestBytes,
	})
	if err != nil {
		mapAnalyzerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, analyzeResponse{
		Spec:           result.Spec,
		Summary:        result.Summary,
		EndpointsCount: result.EndpointsCount,
		Format:         result.Format,
		Language:       result.Language,
		Framework:      result.Framework,
		Model:          result.Model,
		UsedFallback:   result.UsedFallback,
		RetriedSpec:    result.RetriedSpec,
		TokensUsed: tokensUsed{
			Input:  result.InputTokens,
			Output: result.OutputTokens,
		},
		Latency: latencyBreakdown{
			TotalMS:  result.LatencyMS,
			ParserMS: result.ParserMS,
			LLMMS:    result.LLMMS,
			SpecMS:   result.SpecMS,
		},
	})
}
