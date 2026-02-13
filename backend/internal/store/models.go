package store

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	SystemPrompt string         `json:"system_prompt"`
	ModelConfig  map[string]any `json:"model_config"`
	MemoryVector []float32 `json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type Workflow struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	UserID      *uuid.UUID   `json:"user_id,omitempty"`
	Nodes       []NodeConfig `json:"nodes"`
	Edges       []EdgeConfig `json:"edges"`
	Version     int          `json:"version"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type NodeConfig struct {
	ID       string         `json:"id"`
	AgentID  uuid.UUID      `json:"agent_id"`
	Position Position       `json:"position"`
	Data     map[string]any `json:"data,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type EdgeConfig struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type WorkflowNode struct {
	ID              uuid.UUID      `json:"id"`
	WorkflowID      uuid.UUID      `json:"workflow_id"`
	AgentID         uuid.UUID      `json:"agent_id"`
	NodeName        string         `json:"node_name"`
	InputMapping    map[string]any `json:"input_mapping,omitempty"`
	ConfigOverrides map[string]any `json:"config_overrides,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

type Execution struct {
	ID         uuid.UUID      `json:"id"`
	WorkflowID uuid.UUID      `json:"workflow_id"`
	Status     string         `json:"status"`
	Snapshot   map[string]any `json:"snapshot"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ExecutionLog struct {
	ID          uuid.UUID      `json:"id"`
	ExecutionID uuid.UUID      `json:"execution_id"`
	NodeID      *uuid.UUID     `json:"node_id,omitempty"`
	StepType    string         `json:"step_type"`
	Content     map[string]any `json:"content"`
	Sequence    int            `json:"sequence"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Snapshot struct {
	WorkflowID    uuid.UUID      `json:"workflow_id"`
	ExecutionID   uuid.UUID      `json:"execution_id"`
	Nodes         []NodeSnapshot `json:"nodes"`
	Edges         []EdgeConfig   `json:"edges"`
	ExecutionMeta MetaInfo       `json:"execution_meta"`
}

type NodeSnapshot struct {
	NodeID      uuid.UUID `json:"node_id"`
	AgentName   string    `json:"agent_name"`
	Steps       []Step    `json:"steps"`
	FinalOutput string    `json:"final_output"`
}

type Step struct {
	StepID    string         `json:"step_id"`
	Type      string         `json:"type"`
	Input     string         `json:"input,omitempty"`
	Output    string         `json:"output,omitempty"`
	Prompt    string         `json:"prompt,omitempty"`
	Tokens    int            `json:"tokens,omitempty"`
	LatencyMs int64          `json:"latency_ms,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Tool      string         `json:"tool,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Result    string         `json:"result,omitempty"`
}

type MetaInfo struct {
	TotalTokens int     `json:"total_tokens"`
	TotalCost   float64 `json:"total_cost"`
	DurationMs  int64   `json:"duration_ms"`
}
