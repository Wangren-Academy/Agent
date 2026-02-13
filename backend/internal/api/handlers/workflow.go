package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Wangren-Academy/Agent/backend/internal/agent"
	"github.com/Wangren-Academy/Agent/backend/internal/store"
	"github.com/Wangren-Academy/Agent/backend/internal/websocket"
	"github.com/Wangren-Academy/Agent/backend/internal/workflow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WorkflowHandler handles workflow-related requests
type WorkflowHandler struct {
	db       *store.PostgresStore
	hub      *websocket.Hub
	registry *agent.Registry
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(db *store.PostgresStore) *WorkflowHandler {
	return &WorkflowHandler{
		db:       db,
		registry: agent.NewRegistry(),
	}
}

// SetHub sets the websocket hub
func (h *WorkflowHandler) SetHub(hub *websocket.Hub) {
	h.hub = hub
}

// List returns all workflows
func (h *WorkflowHandler) List(c *gin.Context) {
	rows, err := h.db.Pool().Query(context.Background(), `
		SELECT id, name, description, nodes, edges, version, created_at, updated_at
		FROM workflows
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	workflows := []map[string]any{}
	for rows.Next() {
		var (
			id          uuid.UUID
			name        string
			description string
			nodesJSON   []byte
			edgesJSON   []byte
			version     int
			createdAt   time.Time
			updatedAt   time.Time
		)
		err := rows.Scan(&id, &name, &description, &nodesJSON, &edgesJSON, &version, &createdAt, &updatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var nodes []store.NodeConfig
		var edges []store.EdgeConfig
		json.Unmarshal(nodesJSON, &nodes)
		json.Unmarshal(edgesJSON, &edges)

		workflows = append(workflows, map[string]any{
			"id":          id,
			"name":        name,
			"description": description,
			"nodes":       nodes,
			"edges":       edges,
			"version":     version,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		})
	}

	c.JSON(http.StatusOK, workflows)
}

// Create creates a new workflow
func (h *WorkflowHandler) Create(c *gin.Context) {
	var req struct {
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description"`
		Nodes       []store.NodeConfig `json:"nodes"`
		Edges       []store.EdgeConfig `json:"edges"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nodesJSON, _ := json.Marshal(req.Nodes)
	edgesJSON, _ := json.Marshal(req.Edges)

	var id uuid.UUID
	err := h.db.Pool().QueryRow(context.Background(), `
		INSERT INTO workflows (name, description, nodes, edges)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, req.Name, req.Description, nodesJSON, edgesJSON).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          id,
		"name":        req.Name,
		"description": req.Description,
		"nodes":       req.Nodes,
		"edges":       req.Edges,
	})
}

// Get returns a single workflow
func (h *WorkflowHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow id"})
		return
	}

	var (
		name        string
		description string
		nodesJSON   []byte
		edgesJSON   []byte
		version     int
		createdAt   time.Time
		updatedAt   time.Time
	)

	err = h.db.Pool().QueryRow(context.Background(), `
		SELECT name, description, nodes, edges, version, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`, id).Scan(&name, &description, &nodesJSON, &edgesJSON, &version, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	var nodes []store.NodeConfig
	var edges []store.EdgeConfig
	json.Unmarshal(nodesJSON, &nodes)
	json.Unmarshal(edgesJSON, &edges)

	c.JSON(http.StatusOK, map[string]any{
		"id":          id,
		"name":        name,
		"description": description,
		"nodes":       nodes,
		"edges":       edges,
		"version":     version,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
	})
}

// Update updates a workflow
func (h *WorkflowHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow id"})
		return
	}

	var req struct {
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Nodes       []store.NodeConfig `json:"nodes"`
		Edges       []store.EdgeConfig `json:"edges"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nodesJSON, _ := json.Marshal(req.Nodes)
	edgesJSON, _ := json.Marshal(req.Edges)

	_, err = h.db.Pool().Exec(context.Background(), `
		UPDATE workflows
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    nodes = COALESCE($4, nodes),
		    edges = COALESCE($5, edges),
		    version = version + 1
		WHERE id = $1
	`, id, req.Name, req.Description, nodesJSON, edgesJSON)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow updated"})
}

// Delete deletes a workflow
func (h *WorkflowHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow id"})
		return
	}

	_, err = h.db.Pool().Exec(context.Background(), `DELETE FROM workflows WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow deleted"})
}

// Execute starts a workflow execution
func (h *WorkflowHandler) Execute(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow id"})
		return
	}

	var req struct {
		InputData map[string]any `json:"input_data"`
	}
	c.ShouldBindJSON(&req)

	// Create execution record
	executionID := uuid.New()
	_, err = h.db.Pool().Exec(context.Background(), `
		INSERT INTO executions (id, workflow_id, status, snapshot)
		VALUES ($1, $2, 'running', '{}')
	`, executionID, workflowID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get workflow
	var (
		name        string
		description string
		nodesJSON   []byte
		edgesJSON   []byte
	)
	err = h.db.Pool().QueryRow(context.Background(), `
		SELECT name, description, nodes, edges FROM workflows WHERE id = $1
	`, workflowID).Scan(&name, &description, &nodesJSON, &edgesJSON)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "workflow not found"})
		return
	}

	var nodes []store.NodeConfig
	var edges []store.EdgeConfig
	json.Unmarshal(nodesJSON, &nodes)
	json.Unmarshal(edgesJSON, &edges)

	wf := &store.Workflow{
		ID:    workflowID,
		Name:  name,
		Nodes: nodes,
		Edges: edges,
	}

	// Build DAG and scheduler
	dag, err := workflow.NewDAG(wf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentHandler := &AgentHandler{db: h.db}
	scheduler := workflow.NewScheduler(dag, agentHandler, h.registry, executionID)

	// Start execution in background
	go func() {
		ctx := context.Background()
		err := scheduler.Run(ctx, req.InputData)

		status := "success"
		if err != nil {
			status = "failed"
		}

		// Build snapshot
		snapshot := buildSnapshot(workflowID, executionID, scheduler.GetResults(), edges)

		snapshotJSON, _ := json.Marshal(snapshot)
		now := time.Now()
		h.db.Pool().Exec(ctx, `
			UPDATE executions
			SET status = $1, snapshot = $2, finished_at = $3
			WHERE id = $4
		`, status, snapshotJSON, now, executionID)
	}()

	// Stream events via WebSocket
	if h.hub != nil {
		go func() {
			for event := range scheduler.Events() {
				h.hub.BroadcastToExecution(executionID.String(), event.Type, event)
			}
		}()
	}

	c.JSON(http.StatusAccepted, gin.H{
		"execution_id": executionID,
		"status":       "running",
	})
}

func buildSnapshot(workflowID, executionID uuid.UUID, results map[string]*workflow.NodeResult, edges []store.EdgeConfig) store.Snapshot {
	nodeSnapshots := make([]store.NodeSnapshot, 0)
	totalTokens := 0
	var totalDuration int64 = 0

	for nodeID, result := range results {
		nodeSnapshots = append(nodeSnapshots, store.NodeSnapshot{
			NodeID:      uuid.MustParse(nodeID),
			AgentName:   nodeID, // TODO: get actual agent name
			Steps:       result.Steps,
			FinalOutput: result.Output,
		})
		for _, step := range result.Steps {
			totalTokens += step.Tokens
			totalDuration += step.LatencyMs
		}
	}

	return store.Snapshot{
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		Nodes:       nodeSnapshots,
		Edges:       edges,
		ExecutionMeta: store.MetaInfo{
			TotalTokens: totalTokens,
			TotalCost:   float64(totalTokens) * 0.00001, // Rough estimate
			DurationMs:  totalDuration,
		},
	}
}
