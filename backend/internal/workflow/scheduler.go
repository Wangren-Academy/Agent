package workflow

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Wangren-Academy/Agent/backend/internal/agent"
	"github.com/Wangren-Academy/Agent/backend/internal/store"

	"github.com/google/uuid"
)

// Scheduler manages the execution of workflow nodes
type Scheduler struct {
	dag         *DAG
	agentStore  AgentStore
	executor    *agent.Registry
	executionID uuid.UUID

	completed map[string]bool
	results   map[string]*NodeResult
	mu        sync.RWMutex

	eventChan chan ExecutionEvent
	done      chan struct{}
}

// NodeResult stores the result of a node execution
type NodeResult struct {
	NodeID    string
	Output    string
	Steps     []store.Step
	StartTime time.Time
	EndTime   time.Time
	Error     error
}

// ExecutionEvent represents an event during execution
type ExecutionEvent struct {
	Type      string      `json:"type"`
	NodeID    string      `json:"node_id"`
	Step      *store.Step `json:"step,omitempty"`
	Result    *NodeResult `json:"result,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// AgentStore interface for fetching agent configurations
type AgentStore interface {
	GetAgent(ctx context.Context, id uuid.UUID) (*store.Agent, error)
}

// NewScheduler creates a new workflow scheduler
func NewScheduler(dag *DAG, agentStore AgentStore, executor *agent.Registry, executionID uuid.UUID) *Scheduler {
	return &Scheduler{
		dag:         dag,
		agentStore:  agentStore,
		executor:    executor,
		executionID: executionID,
		completed:   make(map[string]bool),
		results:     make(map[string]*NodeResult),
		eventChan:   make(chan ExecutionEvent, 100),
		done:        make(chan struct{}),
	}
}

// Run executes the workflow
func (s *Scheduler) Run(ctx context.Context, input map[string]any) error {
	log.Printf("[Scheduler] Starting execution %s", s.executionID)

	// Initialize with input data
	for nodeID, node := range s.dag.Nodes {
		if len(node.DependsOn) == 0 {
			// Entry node, can use initial input
			s.results[nodeID] = &NodeResult{
				NodeID: nodeID,
				Output: "",
			}
		}
	}

	// Get ready nodes
	readyNodes := s.dag.GetReadyNodes(s.completed)

	var wg sync.WaitGroup

	for _, nodeID := range readyNodes {
		wg.Add(1)
		go func(nid string) {
			defer wg.Done()
			s.executeNode(ctx, nid)
		}(nodeID)
	}

	wg.Wait()
	close(s.eventChan)

	return nil
}

// executeNode executes a single node
func (s *Scheduler) executeNode(ctx context.Context, nodeID string) {
	node := s.dag.Nodes[nodeID]
	startTime := time.Now()

	log.Printf("[Scheduler] Executing node %s (agent: %s)", nodeID, node.AgentID)

	// Get agent configuration
	agentConfig, err := s.agentStore.GetAgent(ctx, node.AgentID)
	if err != nil {
		s.markFailed(nodeID, err, startTime)
		return
	}

	// Get executor
	exec, ok := s.executor.Get(agentConfig.ModelConfig["provider"].(string))
	if !ok {
		s.markFailed(nodeID, fmt.Errorf("no executor for provider: %s", agentConfig.ModelConfig["provider"]), startTime)
		return
	}

	// Build input message from upstream outputs
	input := s.buildInput(nodeID)

	// Build config
	config := agent.Config{
		Provider: agentConfig.ModelConfig["provider"].(string),
		Model:    agentConfig.ModelConfig["model"].(string),
	}
	if temp, ok := agentConfig.ModelConfig["temperature"].(float64); ok {
		config.Temperature = temp
	}

	// Execute
	result, err := exec.Execute(ctx, agent.Message{
		Role:    "user",
		Content: input,
	}, config)

	endTime := time.Now()

	if err != nil {
		s.markFailed(nodeID, err, startTime)
		return
	}

	// Record step
	step := store.Step{
		StepID:    uuid.NewString(),
		Type:      "think",
		Input:     input,
		Output:    result.Content,
		Prompt:    agentConfig.SystemPrompt,
		Tokens:    result.Usage.TotalTokens,
		LatencyMs: result.Latency.Milliseconds(),
		Timestamp: startTime,
	}

	// Send event
	s.eventChan <- ExecutionEvent{
		Type:      "step_complete",
		NodeID:    nodeID,
		Step:      &step,
		Timestamp: time.Now(),
	}

	// Store result
	s.mu.Lock()
	s.completed[nodeID] = true
	s.results[nodeID] = &NodeResult{
		NodeID:    nodeID,
		Output:    result.Content,
		Steps:     []store.Step{step},
		StartTime: startTime,
		EndTime:   endTime,
	}
	s.mu.Unlock()

	log.Printf("[Scheduler] Node %s completed in %v", nodeID, endTime.Sub(startTime))

	// Check for downstream nodes ready to execute
	s.checkDownstream(ctx, nodeID)
}

func (s *Scheduler) buildInput(nodeID string) string {
	node := s.dag.Nodes[nodeID]
	var inputs []string

	for _, depID := range node.DependsOn {
		s.mu.RLock()
		result, ok := s.results[depID]
		s.mu.RUnlock()

		if ok && result != nil {
			inputs = append(inputs, result.Output)
		}
	}

	// Combine inputs
	input := ""
	for i, in := range inputs {
		input += fmt.Sprintf("Input %d: %s\n", i+1, in)
	}
	return input
}

func (s *Scheduler) markFailed(nodeID string, err error, startTime time.Time) {
	s.mu.Lock()
	s.completed[nodeID] = true
	s.results[nodeID] = &NodeResult{
		NodeID:    nodeID,
		Error:     err,
		StartTime: startTime,
		EndTime:   time.Now(),
	}
	s.mu.Unlock()

	s.eventChan <- ExecutionEvent{
		Type:      "node_failed",
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}

	log.Printf("[Scheduler] Node %s failed: %v", nodeID, err)
}

func (s *Scheduler) checkDownstream(ctx context.Context, completedNodeID string) {
	node := s.dag.Nodes[completedNodeID]

	for _, downstreamID := range node.Downstream {
		downstream := s.dag.Nodes[downstreamID]

		// Check if all dependencies are complete
		allComplete := true
		for _, dep := range downstream.DependsOn {
			s.mu.RLock()
			complete := s.completed[dep]
			s.mu.RUnlock()
			if !complete {
				allComplete = false
				break
			}
		}

		if allComplete {
			// Check if already executed
			s.mu.RLock()
			alreadyDone := s.completed[downstreamID]
			s.mu.RUnlock()

			if !alreadyDone {
				go s.executeNode(ctx, downstreamID)
			}
		}
	}
}

// Events returns the event channel
func (s *Scheduler) Events() <-chan ExecutionEvent {
	return s.eventChan
}

// GetResults returns all node results
func (s *Scheduler) GetResults() map[string]*NodeResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results
}

// IsComplete returns true if all nodes are complete
func (s *Scheduler) IsComplete() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.completed) == len(s.dag.Nodes)
}
