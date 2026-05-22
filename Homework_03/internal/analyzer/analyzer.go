package analyzer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shaiso/autodoc-api/internal/llm"
	"github.com/shaiso/autodoc-api/internal/parser"
	"github.com/shaiso/autodoc-api/internal/spec"
)

var (
	ErrInput      = errors.New("invalid input")
	ErrLLM        = errors.New("llm failure")
	ErrSpec       = errors.New("spec build failure")
	ErrTooLarge   = errors.New("input too large")
	ErrTooMany    = errors.New("too many files")
	ErrEmpty      = errors.New("empty input")
	ErrUnknownFmt = errors.New("unknown format")
	ErrLanguage   = errors.New("unknown language")
)

type Request struct {
	Files          []parser.File
	Language       string
	Format         string
	MaxFiles       int
	MaxRequestSize int64
}

type Result struct {
	Spec           string
	Summary        string
	EndpointsCount int
	Format         string
	Language       string
	Framework      string
	Model          string
	UsedFallback   bool
	InputTokens    int
	OutputTokens   int
	LatencyMS      int64
	ParserMS       int64
	LLMMS          int64
	SpecMS         int64
	RetriedSpec    bool
}

type Service struct {
	parser *parser.Service
	llm    llm.Client
	spec   *spec.Service
	logger *slog.Logger
}

func New(p *parser.Service, l llm.Client, s *spec.Service, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{parser: p, llm: l, spec: s, logger: logger}
}

func (s *Service) Analyze(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()

	parserStart := time.Now()
	parsed, err := s.parser.Parse(req.Files, parser.ParseOptions{
		HintLanguage:    req.Language,
		MaxFiles:        req.MaxFiles,
		MaxRequestBytes: req.MaxRequestSize,
	})
	parserMS := time.Since(parserStart).Milliseconds()
	if err != nil {
		return nil, mapParserErr(err)
	}

	llmFiles := make([]llm.FileInput, len(parsed.Files))
	for i, f := range parsed.Files {
		llmFiles[i] = llm.FileInput{Name: f.Name, Content: f.Content}
	}
	userPrompt := llm.BuildUserPrompt(parsed.Language, parsed.Framework, llmFiles)

	llmStart := time.Now()
	llmResp, err := s.llm.Generate(ctx, llm.Request{
		SystemPrompt: llm.SystemPrompt,
		UserPrompt:   userPrompt,
		JSONMode:     true,
	})
	llmMS := time.Since(llmStart).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLLM, err)
	}

	specStart := time.Now()
	specRes, err := s.spec.Build(ctx, llmResp.Content, req.Format)
	retried := false
	if err != nil {
		s.logger.Warn("spec validation failed, retrying once", "error", err.Error())
		retryPrompt := llm.BuildRetryPrompt(llmResp.Content, err.Error())
		retryResp, retryErr := s.llm.Generate(ctx, llm.Request{
			SystemPrompt: llm.SystemPrompt,
			UserPrompt:   retryPrompt,
			JSONMode:     true,
		})
		if retryErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrSpec, err)
		}
		specRes, err = s.spec.Build(ctx, retryResp.Content, req.Format)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSpec, err)
		}
		llmResp.InputTokens += retryResp.InputTokens
		llmResp.OutputTokens += retryResp.OutputTokens
		retried = true
	}
	specMS := time.Since(specStart).Milliseconds()

	return &Result{
		Spec:           specRes.Spec,
		Summary:        specRes.Summary,
		EndpointsCount: specRes.EndpointsCount,
		Format:         specRes.Format,
		Language:       parsed.Language,
		Framework:      parsed.Framework,
		Model:          llmResp.Model,
		UsedFallback:   llmResp.UsedFallback,
		InputTokens:    llmResp.InputTokens,
		OutputTokens:   llmResp.OutputTokens,
		LatencyMS:      time.Since(start).Milliseconds(),
		ParserMS:       parserMS,
		LLMMS:          llmMS,
		SpecMS:         specMS,
		RetriedSpec:    retried,
	}, nil
}

func mapParserErr(err error) error {
	switch {
	case errors.Is(err, parser.ErrEmptyInput):
		return fmt.Errorf("%w: %v", ErrEmpty, err)
	case errors.Is(err, parser.ErrTooManyFiles):
		return fmt.Errorf("%w: %v", ErrTooMany, err)
	case errors.Is(err, parser.ErrTooLarge):
		return fmt.Errorf("%w: %v", ErrTooLarge, err)
	case errors.Is(err, parser.ErrNotSupport):
		return fmt.Errorf("%w: %v", ErrLanguage, err)
	default:
		return fmt.Errorf("%w: %v", ErrInput, err)
	}
}
