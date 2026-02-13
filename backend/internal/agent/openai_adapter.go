package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIAdapter implements Executor for OpenAI models
type OpenAIAdapter struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(apiKey string) *OpenAIAdapter {
	return &OpenAIAdapter{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		baseURL:    "https://api.openai.com/v1",
	}
}

// Name returns the adapter name
func (a *OpenAIAdapter) Name() string {
	return "openai"
}

// Execute sends a request to OpenAI and returns the result
func (a *OpenAIAdapter) Execute(ctx context.Context, input Message, config Config) (*Result, error) {
	startTime := time.Now()

	// Build messages for the OpenAI REST API
	type chatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	messages := []chatMessage{{Role: "user", Content: input.Content}}

	// Build request body
	reqBody := map[string]any{
		"model":       config.Model,
		"messages":    messages,
		"temperature": config.Temperature,
		"max_tokens":  config.MaxTokens,
		"top_p":       config.TopP,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai api error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Response structures
	var apiResp struct {
		Choices []struct {
			Message struct {
				Role         string `json:"role"`
				Content      string `json:"content"`
				FunctionCall *struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function_call,omitempty"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from openai")
	}

	choice := apiResp.Choices[0]
	latency := time.Since(startTime)

	result := &Result{
		Content: choice.Message.Content,
		Usage: TokenUsage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
		Latency: latency,
	}

	// Handle tool calls if present
	// Handle function_call if present
	if choice.Message.FunctionCall != nil {
		var args map[string]any
		if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &args); err != nil {
			args = make(map[string]any)
		}
		result.ToolCalls = []ToolCall{{
			ID:   "",
			Type: "function_call",
			Function: FunctionCall{
				Name:      choice.Message.FunctionCall.Name,
				Arguments: args,
			},
		}}
	}

	return result, nil
}

// AnthropicAdapter implements Executor for Anthropic models
type AnthropicAdapter struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewAnthropicAdapter creates a new Anthropic adapter
func NewAnthropicAdapter(apiKey string) *AnthropicAdapter {
	return &AnthropicAdapter{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		baseURL:    "https://api.anthropic.com/v1",
	}
}

// Name returns the adapter name
func (a *AnthropicAdapter) Name() string {
	return "anthropic"
}

// Execute sends a request to Anthropic and returns the result
func (a *AnthropicAdapter) Execute(ctx context.Context, input Message, config Config) (*Result, error) {
	// TODO: Implement Anthropic API call
	return &Result{
		Content: "Anthropic adapter not yet implemented",
		Usage:   TokenUsage{},
		Latency: 0,
	}, nil
}

// LocalAdapter implements Executor for local models (e.g., Ollama)
type LocalAdapter struct {
	baseURL    string
	httpClient *http.Client
}

// NewLocalAdapter creates a new local model adapter
func NewLocalAdapter(baseURL string) *LocalAdapter {
	return &LocalAdapter{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 300 * time.Second},
	}
}

// Name returns the adapter name
func (a *LocalAdapter) Name() string {
	return "local"
}

// Execute sends a request to a local model and returns the result
func (a *LocalAdapter) Execute(ctx context.Context, input Message, config Config) (*Result, error) {
	// TODO: Implement local model API call (Ollama, vLLM, etc.)
	return &Result{
		Content: "Local adapter not yet implemented",
		Usage:   TokenUsage{},
		Latency: 0,
	}, nil
}
