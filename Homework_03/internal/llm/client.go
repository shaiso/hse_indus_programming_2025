package llm

import (
	"context"
	"errors"
)

var (
	ErrUpstream      = errors.New("llm upstream error")
	ErrTimeout       = errors.New("llm timeout")
	ErrInvalidOutput = errors.New("llm returned invalid output")
)

type Request struct {
	SystemPrompt string
	UserPrompt   string
	JSONMode     bool
}

type Response struct {
	Content      string
	Model        string
	InputTokens  int
	OutputTokens int
	UsedFallback bool
}

type Client interface {
	Generate(ctx context.Context, req Request) (Response, error)
	Ping(ctx context.Context) error
	PrimaryModel() string
}
