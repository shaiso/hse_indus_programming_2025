package parser

import "testing"

func TestFilterRelevant(t *testing.T) {
	in := []File{
		{Name: "main.go", Content: ""},
		{Name: "main_test.go", Content: ""},
		{Name: "vendor/lib/x.go", Content: ""},
		{Name: "node_modules/foo/index.js", Content: ""},
		{Name: "app/server.js", Content: ""},
		{Name: "app/server.test.js", Content: ""},
		{Name: "tests/test_app.py", Content: ""},
		{Name: "src/router.py", Content: ""},
		{Name: "package-lock.json", Content: ""},
		{Name: "migrations/001.sql", Content: ""},
		{Name: "static/img.png", Content: ""},
	}
	out := FilterRelevant(in)
	wantNames := map[string]bool{
		"main.go":       true,
		"app/server.js": true,
		"src/router.py": true,
	}
	if len(out) != len(wantNames) {
		t.Fatalf("expected %d files, got %d (%v)", len(wantNames), len(out), names(out))
	}
	for _, f := range out {
		if !wantNames[f.Name] {
			t.Errorf("unexpected file kept: %s", f.Name)
		}
	}
}

func TestParseLimits(t *testing.T) {
	s := New()

	t.Run("empty", func(t *testing.T) {
		_, err := s.Parse(nil, ParseOptions{})
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("too many files", func(t *testing.T) {
		files := make([]File, 5)
		for i := range files {
			files[i] = File{Name: "a.go", Content: "package main"}
		}
		_, err := s.Parse(files, ParseOptions{MaxFiles: 3})
		if err == nil {
			t.Fatal("expected error for too many files")
		}
	})

	t.Run("too large", func(t *testing.T) {
		big := make([]byte, 100)
		_, err := s.Parse(
			[]File{{Name: "a.go", Content: string(big)}},
			ParseOptions{MaxRequestBytes: 50},
		)
		if err == nil {
			t.Fatal("expected error for size limit")
		}
	})

	t.Run("hint language wins", func(t *testing.T) {
		out, err := s.Parse(
			[]File{{Name: "x.unknown", Content: "code"}},
			ParseOptions{HintLanguage: "go"},
		)
		if err != nil {
			t.Fatal(err)
		}
		if out.Language != "go" {
			t.Fatalf("expected go, got %q", out.Language)
		}
	})
}

func names(files []File) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Name
	}
	return out
}
