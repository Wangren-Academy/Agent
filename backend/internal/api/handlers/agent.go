package handlers

import (
	"context"
	"net/http"

	"github.com/Wangren-Academy/Agent/backend/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AgentHandler handles agent-related requests
type AgentHandler struct {
	db *store.PostgresStore
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(db *store.PostgresStore) *AgentHandler {
	return &AgentHandler{db: db}
}

// List returns all agents
func (h *AgentHandler) List(c *gin.Context) {
	rows, err := h.db.Pool().Query(context.Background(), `
		SELECT id, name, description, system_prompt, model_config, created_at, updated_at
		FROM agents
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	agents := []store.Agent{}
	for rows.Next() {
		var a store.Agent
		err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.ModelConfig, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		agents = append(agents, a)
	}

	c.JSON(http.StatusOK, agents)
}

// Create creates a new agent
func (h *AgentHandler) Create(c *gin.Context) {
	var req struct {
		Name         string         `json:"name" binding:"required"`
		Description  string         `json:"description"`
		SystemPrompt string         `json:"system_prompt" binding:"required"`
		ModelConfig  map[string]any `json:"model_config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ModelConfig == nil {
		req.ModelConfig = map[string]any{
			"provider":    "openai",
			"model":       "gpt-4",
			"temperature": 0.7,
		}
	}

	var id uuid.UUID
	err := h.db.Pool().QueryRow(context.Background(), `
		INSERT INTO agents (name, description, system_prompt, model_config)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, req.Name, req.Description, req.SystemPrompt, req.ModelConfig).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            id,
		"name":          req.Name,
		"description":   req.Description,
		"system_prompt": req.SystemPrompt,
		"model_config":  req.ModelConfig,
	})
}

// Get returns a single agent
func (h *AgentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent id"})
		return
	}

	var a store.Agent
	err = h.db.Pool().QueryRow(context.Background(), `
		SELECT id, name, description, system_prompt, model_config, created_at, updated_at
		FROM agents
		WHERE id = $1
	`, id).Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.ModelConfig, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, a)
}

// Update updates an agent
func (h *AgentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent id"})
		return
	}

	var req struct {
		Name         string         `json:"name"`
		Description  string         `json:"description"`
		SystemPrompt string         `json:"system_prompt"`
		ModelConfig  map[string]any `json:"model_config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.db.Pool().Exec(context.Background(), `
		UPDATE agents
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    system_prompt = COALESCE($4, system_prompt),
		    model_config = COALESCE($5, model_config)
		WHERE id = $1
	`, id, req.Name, req.Description, req.SystemPrompt, req.ModelConfig)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent updated"})
}

// Delete deletes an agent
func (h *AgentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent id"})
		return
	}

	_, err = h.db.Pool().Exec(context.Background(), `DELETE FROM agents WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
}

// GetAgent retrieves an agent by ID (for AgentStore interface)
func (h *AgentHandler) GetAgent(ctx context.Context, id uuid.UUID) (*store.Agent, error) {
	var a store.Agent
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, name, description, system_prompt, model_config, created_at, updated_at
		FROM agents
		WHERE id = $1
	`, id).Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.ModelConfig, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &a, nil
}
