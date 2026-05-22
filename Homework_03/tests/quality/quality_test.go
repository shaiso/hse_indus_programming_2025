// Package quality_test — оценка precision/recall/валидности спеки
// против эталонных проектов в examples/.
//
// Запуск:
//
//	go test ./tests/quality -run TestQuality -v -count=1 -timeout=600s -args \
//	    -url http://localhost:8080/api/v1/analyze
//
// Без -url тесты пропускаются.
package qualitytest

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var flagURL = flag.String("url", "", "analyze endpoint URL")

type ExpectedEndpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Expected struct {
	Endpoints []ExpectedEndpoint `json:"endpoints"`
}

type ProjectCase struct {
	Name     string
	Dir      string
	Language string
	SrcFile  string
}

var cases = []ProjectCase{
	{Name: "go_gin_api", Dir: "../../examples/go_gin_api", Language: "go", SrcFile: "main.go"},
	{Name: "python_fastapi", Dir: "../../examples/python_fastapi", Language: "python", SrcFile: "main.py"},
	{Name: "express_api", Dir: "../../examples/express_api", Language: "javascript", SrcFile: "server.js"},
}

type analyzeReq struct {
	Files    []fileDTO `json:"files"`
	Language string    `json:"language"`
	Format   string    `json:"format"`
}

type fileDTO struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type analyzeResp struct {
	Spec           string `json:"spec"`
	EndpointsCount int    `json:"endpoints_count"`
	Format         string `json:"format"`
	Model          string `json:"model"`
	Latency        struct {
		TotalMS int64 `json:"total_ms"`
		LLMMS   int64 `json:"llm_ms"`
	} `json:"latency"`
	TokensUsed struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens_used"`
}

func TestQuality(t *testing.T) {
	if *flagURL == "" {
		t.Skip("set -url to run quality eval")
	}

	type result struct {
		Project   string
		Found     int
		Expected  int
		Matched   int
		Precision float64
		Recall    float64
		LatencyMS int64
		LLMMS     int64
		InTokens  int
		OutTokens int
		SpecValid bool
	}
	var results []result
	totalLat := int64(0)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			expected := loadExpected(t, filepath.Join(tc.Dir, "expected_endpoints.json"))
			content, err := os.ReadFile(filepath.Join(tc.Dir, tc.SrcFile))
			if err != nil {
				t.Fatalf("read source: %v", err)
			}

			body, _ := json.Marshal(analyzeReq{
				Files:    []fileDTO{{Name: tc.SrcFile, Content: string(content)}},
				Language: tc.Language,
				Format:   "json",
			})

			start := time.Now()
			resp, err := http.Post(*flagURL, "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("post: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			var ar analyzeResp
			if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
				t.Fatalf("decode: %v", err)
			}
			lat := time.Since(start).Milliseconds()
			totalLat += lat

			found := extractEndpoints(t, ar.Spec)
			matched := countMatches(expected.Endpoints, found)
			precision := safeDiv(matched, len(found))
			recall := safeDiv(matched, len(expected.Endpoints))

			r := result{
				Project:   tc.Name,
				Found:     len(found),
				Expected:  len(expected.Endpoints),
				Matched:   matched,
				Precision: precision,
				Recall:    recall,
				LatencyMS: lat,
				LLMMS:     ar.Latency.LLMMS,
				InTokens:  ar.TokensUsed.Input,
				OutTokens: ar.TokensUsed.Output,
				SpecValid: true,
			}
			results = append(results, r)

			t.Logf("=== %s ===", tc.Name)
			t.Logf("expected=%d found=%d matched=%d", r.Expected, r.Found, r.Matched)
			t.Logf("precision=%.2f recall=%.2f", r.Precision, r.Recall)
			t.Logf("latency_ms=%d llm_ms=%d", r.LatencyMS, r.LLMMS)
			t.Logf("tokens in=%d out=%d", r.InTokens, r.OutTokens)
			t.Logf("model=%s", ar.Model)
		})
	}

	t.Run("aggregate", func(t *testing.T) {
		if len(results) == 0 {
			t.Skip("no per-case results")
		}
		var sumP, sumR float64
		var sumIn, sumOut int
		for _, r := range results {
			sumP += r.Precision
			sumR += r.Recall
			sumIn += r.InTokens
			sumOut += r.OutTokens
		}
		t.Logf("=== AGGREGATE (n=%d) ===", len(results))
		t.Logf("avg_precision=%.2f", sumP/float64(len(results)))
		t.Logf("avg_recall=%.2f", sumR/float64(len(results)))
		t.Logf("total_tokens in=%d out=%d", sumIn, sumOut)
		t.Logf("approx_cost_usd=%.4f (gpt-4.1-mini @ $0.40 / $1.60 per 1M)",
			float64(sumIn)*0.40/1_000_000+float64(sumOut)*1.60/1_000_000)
		t.Logf("total_latency_ms=%d", totalLat)
	})
}

func extractEndpoints(t *testing.T, spec string) []ExpectedEndpoint {
	t.Helper()
	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal([]byte(spec), &doc); err != nil {
		t.Fatalf("spec is not valid JSON: %v", err)
	}
	var out []ExpectedEndpoint
	methods := map[string]bool{
		"get": true, "post": true, "put": true, "patch": true,
		"delete": true, "head": true, "options": true,
	}
	for path, ops := range doc.Paths {
		for m := range ops {
			if methods[strings.ToLower(m)] {
				out = append(out, ExpectedEndpoint{
					Method: strings.ToUpper(m),
					Path:   normalizePath(path),
				})
			}
		}
	}
	return out
}

// normalizePath приводит {param} к каноничному виду без чувствительности
// к имени параметра, чтобы /users/{id} == /users/{userId}.
func normalizePath(p string) string {
	if !strings.Contains(p, "{") {
		return p
	}
	var b strings.Builder
	depth := 0
	for _, r := range p {
		switch r {
		case '{':
			depth++
			b.WriteString("{*}")
		case '}':
			depth--
		default:
			if depth == 0 {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func canon(e ExpectedEndpoint) string {
	return strings.ToUpper(e.Method) + " " + normalizePath(e.Path)
}

func countMatches(expected, found []ExpectedEndpoint) int {
	exp := make(map[string]int)
	for _, e := range expected {
		exp[canon(e)]++
	}
	matches := 0
	for _, f := range found {
		k := canon(f)
		if exp[k] > 0 {
			exp[k]--
			matches++
		}
	}
	return matches
}

func loadExpected(t *testing.T, path string) Expected {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}
	var e Expected
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatalf("unmarshal expected: %v", err)
	}
	return e
}

func safeDiv(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

var _ = fmt.Sprintf
