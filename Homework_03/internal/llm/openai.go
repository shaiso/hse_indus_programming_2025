package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	APIKey        string
	BaseURL       string
	Model         string
	FallbackModel string
	Timeout       time.Duration
	Logger        *slog.Logger
}

type OpenAIClient struct {
	client        *openai.Client
	model         string
	fallbackModel string
	timeout       time.Duration
	logger        *slog.Logger
}

func NewOpenAI(cfg OpenAIConfig) *OpenAIClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	clientCfg := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientCfg.BaseURL = cfg.BaseURL
	}

	return &OpenAIClient{
		client:        openai.NewClientWithConfig(clientCfg),
		model:         cfg.Model,
		fallbackModel: cfg.FallbackModel,
		timeout:       cfg.Timeout,
		logger:        cfg.Logger,
	}
}

func (c *OpenAIClient) PrimaryModel() string { return c.model }

func (c *OpenAIClient) Generate(ctx context.Context, req Request) (Response, error) {
	resp, err := c.callModel(ctx, c.model, req, false)
	if err == nil {
		return resp, nil
	}
	if !c.shouldFallback(err) || c.fallbackModel == "" || c.fallbackModel == c.model {
		return Response{}, err
	}
	c.logger.Warn("primary model failed, retrying with fallback",
		"primary", c.model,
		"fallback", c.fallbackModel,
		"error", err.Error(),
	)
	return c.callModel(ctx, c.fallbackModel, req, true)
}

func (c *OpenAIClient) callModel(ctx context.Context, model string, req Request, isFallback bool) (Response, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	chatReq := openai.ChatCompletionRequest{
		Model:       model,
		Temperature: 0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: req.SystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: req.UserPrompt},
		},
	}
	if req.JSONMode {
		chatReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	resp, err := c.client.CreateChatCompletion(callCtx, chatReq)
	if err != nil {
		if errors.Is(callCtx.Err(), context.DeadlineExceeded) {
			return Response{}, fmt.Errorf("%w: %s", ErrTimeout, model)
		}
		return Response{}, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	if len(resp.Choices) == 0 {
		return Response{}, fmt.Errorf("%w: empty choices", ErrInvalidOutput)
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return Response{}, fmt.Errorf("%w: empty content", ErrInvalidOutput)
	}

	return Response{
		Content:      content,
		Model:        model,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		UsedFallback: isFallback,
	}, nil
}

func (c *OpenAIClient) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := c.client.ListModels(pingCtx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	return nil
}

func (c *OpenAIClient) shouldFallback(err error) bool {
	return errors.Is(err, ErrUpstream) || errors.Is(err, ErrTimeout) || errors.Is(err, ErrInvalidOutput)
}
