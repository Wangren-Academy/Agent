package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Wangren-Academy/Agent/backend/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExecutionHandler handles execution-related requests
type ExecutionHandler struct {
	db *store.PostgresStore
}

// NewExecutionHandler creates a new execution handler
func NewExecutionHandler(db *store.PostgresStore) *ExecutionHandler {
	return &ExecutionHandler{db: db}
}

// List returns all executions
func (h *ExecutionHandler) List(c *gin.Context) {
	workflowID := c.Query("workflow_id")
	status := c.Query("status")

	query := `
		SELECT id, workflow_id, status, started_at, finished_at, created_at
		FROM executions
	`
	args := []any{}
	argIdx := 1

	if workflowID != "" {
		query += " WHERE workflow_id = $" + string(rune('0'+argIdx))
		id, err := uuid.Parse(workflowID)
		if err == nil {
			args = append(args, id)
			argIdx++
		}
	}

	if status != "" {
		if len(args) > 0 {
			query += " AND status = $" + string(rune('0'+argIdx))
		} else {
			query += " WHERE status = $" + string(rune('0'+argIdx))
		}
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := h.db.Pool().Query(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	executions := []map[string]any{}
	for rows.Next() {
		var (
			id         uuid.UUID
			wfID       uuid.UUID
			status     string
			startedAt  time.Time
			finishedAt *time.Time
			createdAt  time.Time
		)
		err := rows.Scan(&id, &wfID, &status, &startedAt, &finishedAt, &createdAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		executions = append(executions, map[string]any{
			"id":          id,
			"workflow_id": wfID,
			"status":      status,
			"started_at":  startedAt,
			"finished_at": finishedAt,
			"created_at":  createdAt,
		})
	}

	c.JSON(http.StatusOK, executions)
}

// Get returns a single execution with full snapshot
func (h *ExecutionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution id"})
		return
	}

	var (
		workflowID   uuid.UUID
		status       string
		snapshotJSON []byte
		startedAt    time.Time
		finishedAt   *time.Time
		createdAt    time.Time
	)

	err = h.db.Pool().QueryRow(context.Background(), `
		SELECT workflow_id, status, snapshot, started_at, finished_at, created_at
		FROM executions
		WHERE id = $1
	`, id).Scan(&workflowID, &status, &snapshotJSON, &startedAt, &finishedAt, &createdAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	var snapshot store.Snapshot
	json.Unmarshal(snapshotJSON, &snapshot)

	c.JSON(http.StatusOK, map[string]any{
		"id":          id,
		"workflow_id": workflowID,
		"status":      status,
		"snapshot":    snapshot,
		"started_at":  startedAt,
		"finished_at": finishedAt,
		"created_at":  createdAt,
	})
}

// Replay handles sandbox replay requests
func (h *ExecutionHandler) Replay(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution id"})
		return
	}

	var req struct {
		ModifiedSteps []struct {
			StepID    string `json:"step_id"`
			NewOutput string `json:"new_output"`
		} `json:"modified_steps"`
	}
	c.ShouldBindJSON(&req)

	// Get original execution
	var snapshotJSON []byte
	err = h.db.Pool().QueryRow(context.Background(), `
		SELECT snapshot FROM executions WHERE id = $1
	`, id).Scan(&snapshotJSON)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	var snapshot store.Snapshot
	json.Unmarshal(snapshotJSON, &snapshot)

	// Apply modifications to snapshot
	modificationMap := make(map[string]string)
	for _, mod := range req.ModifiedSteps {
		modificationMap[mod.StepID] = mod.NewOutput
	}

	for i, node := range snapshot.Nodes {
		for j, step := range node.Steps {
			if newOutput, ok := modificationMap[step.StepID]; ok {
				snapshot.Nodes[i].Steps[j].Output = newOutput
			}
		}
	}

	// Create new execution for replay
	newExecutionID := uuid.New()
	workflowID := snapshot.WorkflowID

	_, err = h.db.Pool().Exec(context.Background(), `
		INSERT INTO executions (id, workflow_id, status, snapshot)
		VALUES ($1, $2, 'replaying', $3)
	`, newExecutionID, workflowID, snapshotJSON)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: Trigger re-execution from modified state

	c.JSON(http.StatusAccepted, gin.H{
		"original_execution_id": id,
		"new_execution_id":      newExecutionID,
		"status":                "replaying",
		"modifications_applied": len(req.ModifiedSteps),
	})
}
