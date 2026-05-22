// This directory is an isolated fixture used by tests/quality to exercise
// the parser against a real Gin project. It is intentionally a separate
// Go module so its dependencies (gin) do not leak into the main module.

module example.com/autodoc/fixtures/go_gin_api

go 1.25
