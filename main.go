package main

import (
	"fmt"
	"log"

	"cosine/config"
	"cosine/database"
	"cosine/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Register routes
	r.GET("/health", handlers.HealthHandler)
	r.GET("/v1/models", handlers.ModelsHandler)
	r.POST("/v1/chat/completions", handlers.ChatCompletionsHandler)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting cosine2api server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
