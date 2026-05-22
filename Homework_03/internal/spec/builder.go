package spec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

const (
	FormatYAML = "yaml"
	FormatJSON = "json"
)

var (
	ErrEmptyInput = errors.New("empty spec")
	ErrInvalidFmt = errors.New("invalid format")
)

type Result struct {
	Spec           string
	Format         string
	EndpointsCount int
	Summary        string
}

type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) Build(ctx context.Context, raw, format string) (*Result, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrEmptyInput
	}
	raw = stripCodeFences(raw)

	doc, err := loadAndValidate(ctx, raw)
	if err != nil {
		return nil, err
	}

	endpoints := countEndpoints(doc)
	summary := buildSummary(doc, endpoints)

	out, normFmt, err := serialize(doc, format)
	if err != nil {
		return nil, err
	}
	return &Result{
		Spec:           out,
		Format:         normFmt,
		EndpointsCount: endpoints,
		Summary:        summary,
	}, nil
}

func loadAndValidate(ctx context.Context, raw string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false

	doc, err := loader.LoadFromData([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("parse openapi: %w", err)
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validate openapi: %w", err)
	}
	return doc, nil
}

func countEndpoints(doc *openapi3.T) int {
	if doc.Paths == nil {
		return 0
	}
	count := 0
	for _, item := range doc.Paths.Map() {
		if item == nil {
			continue
		}
		for _, op := range item.Operations() {
			if op != nil {
				count++
			}
		}
	}
	return count
}

func buildSummary(doc *openapi3.T, endpoints int) string {
	title := "API"
	if doc.Info != nil && doc.Info.Title != "" {
		title = doc.Info.Title
	}
	return fmt.Sprintf("%s — detected %d endpoint(s)", title, endpoints)
}

func serialize(doc *openapi3.T, format string) (string, string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = FormatYAML
	}
	switch format {
	case FormatJSON:
		b, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return "", "", err
		}
		return string(b), FormatJSON, nil
	case FormatYAML:
		// kin-openapi's MarshalJSON is canonical — round-trip into yaml.
		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			return "", "", err
		}
		var generic any
		if err := json.Unmarshal(jsonBytes, &generic); err != nil {
			return "", "", err
		}
		var buf strings.Builder
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(generic); err != nil {
			return "", "", err
		}
		_ = enc.Close()
		return buf.String(), FormatYAML, nil
	default:
		return "", "", fmt.Errorf("%w: %q", ErrInvalidFmt, format)
	}
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return s
	}
	// drop opening fence
	lines = lines[1:]
	// drop closing fence if present
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
