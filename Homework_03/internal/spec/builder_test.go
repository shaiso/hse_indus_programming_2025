package spec

import (
	"context"
	"strings"
	"testing"
)

const validSpec = `{
  "openapi": "3.0.3",
  "info": {"title": "Demo", "version": "1.0.0"},
  "paths": {
    "/users": {
      "get": {
        "summary": "List users",
        "responses": {
          "200": {"description": "OK"}
        }
      },
      "post": {
        "summary": "Create user",
        "responses": {
          "201": {"description": "Created"}
        }
      }
    },
    "/users/{id}": {
      "get": {
        "summary": "Get user",
        "parameters": [
          {"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}
        ],
        "responses": {
          "200": {"description": "OK"}
        }
      }
    }
  }
}`

const invalidSpec = `{"openapi": "3.0.3", "info": {"title": "x"}}`

func TestBuild_ValidYAML(t *testing.T) {
	s := New()
	r, err := s.Build(context.Background(), validSpec, "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Format != FormatYAML {
		t.Errorf("format = %q, want yaml", r.Format)
	}
	if r.EndpointsCount != 3 {
		t.Errorf("endpoints = %d, want 3", r.EndpointsCount)
	}
	if !strings.Contains(r.Spec, "openapi:") {
		t.Errorf("expected YAML output, got: %s", r.Spec)
	}
}

func TestBuild_ValidJSON(t *testing.T) {
	s := New()
	r, err := s.Build(context.Background(), validSpec, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Format != FormatJSON {
		t.Errorf("format = %q, want json", r.Format)
	}
	if !strings.HasPrefix(strings.TrimSpace(r.Spec), "{") {
		t.Errorf("expected JSON output, got: %s", r.Spec)
	}
}

func TestBuild_InvalidSpec(t *testing.T) {
	s := New()
	_, err := s.Build(context.Background(), invalidSpec, "yaml")
	if err == nil {
		t.Fatal("expected error for invalid spec (missing version)")
	}
}

func TestBuild_DefaultFormat(t *testing.T) {
	s := New()
	r, err := s.Build(context.Background(), validSpec, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Format != FormatYAML {
		t.Errorf("default format = %q, want yaml", r.Format)
	}
}

func TestBuild_StripsCodeFences(t *testing.T) {
	s := New()
	wrapped := "```json\n" + validSpec + "\n```"
	r, err := s.Build(context.Background(), wrapped, "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.EndpointsCount != 3 {
		t.Errorf("endpoints = %d, want 3", r.EndpointsCount)
	}
}

func TestBuild_Empty(t *testing.T) {
	s := New()
	_, err := s.Build(context.Background(), "   ", "yaml")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestBuild_UnknownFormat(t *testing.T) {
	s := New()
	_, err := s.Build(context.Background(), validSpec, "xml")
	if err == nil {
		t.Fatal("expected error for xml format")
	}
}
