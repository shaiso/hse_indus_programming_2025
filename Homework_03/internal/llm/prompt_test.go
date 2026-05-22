package llm

import (
	"strings"
	"testing"
)

func TestBuildUserPrompt(t *testing.T) {
	p := BuildUserPrompt("go", "gin", []FileInput{
		{Name: "main.go", Content: "package main\nfunc main() {}"},
	})

	if !strings.Contains(p, "Detected language: go") {
		t.Errorf("missing language hint")
	}
	if !strings.Contains(p, "Detected framework: gin") {
		t.Errorf("missing framework hint")
	}
	if !strings.Contains(p, "===== FILE: main.go =====") {
		t.Errorf("missing file header")
	}
	if !strings.Contains(p, "package main") {
		t.Errorf("missing file content")
	}
	if !strings.Contains(p, "OpenAPI") {
		t.Errorf("missing instruction to return OpenAPI")
	}
}

func TestBuildUserPrompt_NoLangHint(t *testing.T) {
	p := BuildUserPrompt("", "", []FileInput{{Name: "x.py", Content: "pass"}})
	if strings.Contains(p, "Detected language") {
		t.Errorf("should not include lang line when empty")
	}
}

func TestBuildRetryPrompt(t *testing.T) {
	prev := strings.Repeat("a", 3000)
	p := BuildRetryPrompt(prev, "missing version")
	if !strings.Contains(p, "missing version") {
		t.Errorf("retry prompt should include validation error")
	}
	if !strings.Contains(p, "[truncated]") {
		t.Errorf("retry prompt should truncate long previous output")
	}
}
