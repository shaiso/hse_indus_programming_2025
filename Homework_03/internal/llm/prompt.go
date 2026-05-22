package llm

import (
	"fmt"
	"strings"
)

const SystemPrompt = `You are an expert API architect who reverse-engineers OpenAPI 3.0 specifications from source code.

Rules:
1. Output VALID OpenAPI 3.0 JSON only. No prose. No markdown fences.
2. Top-level fields: openapi (must be "3.0.3"), info (with title and version), paths, components.
3. For every detected HTTP route emit a path item with the correct method, parameters, request body schema, and response schemas.
4. Infer parameter types and required flags from validation code, struct tags, and pydantic models.
5. If the same model is reused, place it under components.schemas and reference via $ref.
6. Do NOT invent endpoints that are not present in the code.
7. If you are uncertain about a field, omit it rather than guess.
8. Use idiomatic OpenAPI: tags, summaries, descriptions where the code provides them.`

func BuildUserPrompt(language, framework string, files []FileInput) string {
	var b strings.Builder
	b.WriteString("Generate an OpenAPI 3.0 specification for the following service.\n\n")
	if language != "" {
		fmt.Fprintf(&b, "Detected language: %s\n", language)
	}
	if framework != "" {
		fmt.Fprintf(&b, "Detected framework: %s\n", framework)
	}
	b.WriteString("\nSource files:\n")
	for _, f := range files {
		fmt.Fprintf(&b, "\n===== FILE: %s =====\n", f.Name)
		b.WriteString(f.Content)
		if !strings.HasSuffix(f.Content, "\n") {
			b.WriteString("\n")
		}
	}
	b.WriteString("\nReturn ONLY the OpenAPI JSON document.")
	return b.String()
}

type FileInput struct {
	Name    string
	Content string
}

func BuildRetryPrompt(previousOutput, validationError string) string {
	return fmt.Sprintf(`Your previous response failed OpenAPI 3.0 validation.

Validation error:
%s

Previous output (first 2000 chars):
%s

Fix the spec and return ONLY the corrected OpenAPI 3.0 JSON document.`,
		validationError, truncate(previousOutput, 2000))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
