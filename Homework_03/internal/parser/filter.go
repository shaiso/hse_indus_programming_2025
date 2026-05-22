package parser

import (
	"path/filepath"
	"strings"
)

func FilterRelevant(files []File) []File {
	relevant := make([]File, 0, len(files))
	for _, f := range files {
		if isIrrelevant(f.Name) {
			continue
		}
		relevant = append(relevant, f)
	}
	return relevant
}

func isIrrelevant(name string) bool {
	low := strings.ToLower(name)
	base := strings.ToLower(filepath.Base(name))

	// Tests
	if strings.HasSuffix(low, "_test.go") ||
		strings.HasSuffix(low, ".test.js") ||
		strings.HasSuffix(low, ".test.ts") ||
		strings.HasSuffix(low, ".spec.js") ||
		strings.HasSuffix(low, ".spec.ts") {
		return true
	}
	if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
		return true
	}

	// Build artefacts and metadata
	switch base {
	case "package-lock.json", "yarn.lock", "go.sum", "poetry.lock", "pipfile.lock":
		return true
	}

	// Dotfiles, vendor, virtual envs, build dirs, migrations, static assets
	for p := range strings.SplitSeq(strings.ReplaceAll(low, "\\", "/"), "/") {
		switch p {
		case "vendor", "node_modules", "dist", "build", ".git", "__pycache__",
			".venv", "venv", "env", "migrations", "static", "assets":
			return true
		}
	}

	return false
}
