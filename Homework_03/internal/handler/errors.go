package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shaiso/autodoc-api/internal/analyzer"
	"github.com/shaiso/autodoc-api/internal/llm"
)

type errorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code, msg, details string) {
	writeJSON(w, status, errorResponse{Error: msg, Code: code, Details: details})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func mapAnalyzerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, analyzer.ErrEmpty):
		writeError(w, http.StatusBadRequest, "empty_input", "request must contain at least one file", err.Error())
	case errors.Is(err, analyzer.ErrTooMany), errors.Is(err, analyzer.ErrTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, "input_too_large", "input exceeds limits, split the project", err.Error())
	case errors.Is(err, analyzer.ErrLanguage):
		writeError(w, http.StatusUnprocessableEntity, "unsupported_language", "language is not supported", err.Error())
	case errors.Is(err, analyzer.ErrInput):
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid input", err.Error())
	case errors.Is(err, analyzer.ErrSpec):
		writeError(w, http.StatusBadGateway, "spec_invalid", "LLM produced invalid OpenAPI specification", err.Error())
	case errors.Is(err, analyzer.ErrLLM):
		status := http.StatusBadGateway
		code := "llm_upstream"
		msg := "LLM provider error"
		if errors.Is(err, llm.ErrTimeout) {
			status = http.StatusGatewayTimeout
			code = "llm_timeout"
			msg = "LLM request timed out"
		}
		writeError(w, status, code, msg, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal", "internal server error", err.Error())
	}
}
