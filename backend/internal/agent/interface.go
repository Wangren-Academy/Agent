package agent

import (
	"context"
	"time"
)

// Executor defines the interface for AI model adapters
type Executor interface {
	Execute(ctx context.Context, input Message, config Config) (*Result, error)
	Name() string
}

// Message represents a single message in the conversation
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool/function call
type ToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function FunctionCall   `json:"function"`
}

// FunctionCall represents a function call details
type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Result represents the execution result from an AI model
type Result struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     TokenUsage `json:"usage"`
	Latency   time.Duration `json:"latency"`
}

// TokenUsage represents token consumption
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Config represents model configuration
type Config struct {
	Provider    string         `json:"provider"`
	Model       string         `json:"model"`
	Temperature float64        `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	Tools       []Tool         `json:"tools,omitempty"`
	Extra       map[string]any `json:"extra,omitempty"`
}

// Tool represents a tool/function definition
type Tool struct {
	Type     string       `json:"type"`
	Function FunctionDef  `json:"function"`
}

// FunctionDef represents a function definition
type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Registry manages all available agent executors
type Registry struct {
	executors map[string]Executor
}

// NewRegistry creates a new executor registry
func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[string]Executor),
	}
}

// Register adds an executor to the registry
func (r *Registry) Register(e Executor) {
	r.executors[e.Name()] = e
}

// Get retrieves an executor by name
func (r *Registry) Get(name string) (Executor, bool) {
	e, ok := r.executors[name]
	return e, ok
}
