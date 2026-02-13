package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Wangren-Academy/Agent/backend/internal/api/handlers"
	"github.com/Wangren-Academy/Agent/backend/internal/store"
	"github.com/Wangren-Academy/Agent/backend/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		getEnv("DB_USER", "agent"),
		getEnv("DB_PASSWORD", "secret"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "agentforge"),
	)

	db, err := store.NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Setup Gin router
	gin.SetMode(getEnv("GIN_MODE", "debug"))
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Agent routes
		agentHandler := handlers.NewAgentHandler(db)
		api.GET("/agents", agentHandler.List)
		api.POST("/agents", agentHandler.Create)
		api.GET("/agents/:id", agentHandler.Get)
		api.PUT("/agents/:id", agentHandler.Update)
		api.DELETE("/agents/:id", agentHandler.Delete)

		// Workflow routes
		workflowHandler := handlers.NewWorkflowHandler(db)
		api.GET("/workflows", workflowHandler.List)
		api.POST("/workflows", workflowHandler.Create)
		api.GET("/workflows/:id", workflowHandler.Get)
		api.PUT("/workflows/:id", workflowHandler.Update)
		api.DELETE("/workflows/:id", workflowHandler.Delete)
		api.POST("/workflows/:id/execute", workflowHandler.Execute)

		// Execution routes
		executionHandler := handlers.NewExecutionHandler(db)
		api.GET("/executions", executionHandler.List)
		api.GET("/executions/:id", executionHandler.Get)
		api.POST("/executions/:id/replay", executionHandler.Replay)
	}

	// WebSocket endpoint
	r.GET("/ws/executions/:id", func(c *gin.Context) {
		websocket.ServeWS(hub, c.Writer, c.Request)
	})

	// Start server
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 AgentForge Backend running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
