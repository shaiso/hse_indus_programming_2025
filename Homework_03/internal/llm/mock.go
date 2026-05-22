package llm

import (
	"context"
	"errors"
	"sync"
)

type MockClient struct {
	mu       sync.Mutex
	Response string
	Err      error
	Calls    int
	PingErr  error
}

func NewMock() *MockClient {
	return &MockClient{
		Response: defaultMockSpec,
	}
}

func (m *MockClient) Generate(_ context.Context, _ Request) (Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls++
	if m.Err != nil {
		return Response{}, m.Err
	}
	return Response{
		Content:      m.Response,
		Model:        "mock",
		InputTokens:  100,
		OutputTokens: 200,
	}, nil
}

func (m *MockClient) Ping(_ context.Context) error {
	if m.PingErr != nil {
		return m.PingErr
	}
	return nil
}

func (m *MockClient) PrimaryModel() string { return "mock" }

func (m *MockClient) SetResponse(s string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Response = s
}

func (m *MockClient) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Err = err
}

func (m *MockClient) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Calls
}

var ErrMockNotConfigured = errors.New("mock not configured")

const defaultMockSpec = `{
  "openapi": "3.0.3",
  "info": {"title": "Mock API", "version": "0.1.0"},
  "paths": {
    "/ping": {
      "get": {
        "summary": "Ping",
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {"type": "object", "properties": {"message": {"type": "string"}}}
              }
            }
          }
        }
      }
    }
  }
}`
