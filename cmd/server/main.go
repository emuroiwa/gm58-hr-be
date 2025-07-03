package main

import (
	"gm58-hr-backend/internal/api/routes"
	"gm58-hr-backend/internal/config"
	"gm58-hr-backend/internal/database"
	"gm58-hr-backend/pkg/logger"
	"gm58-hr-backend/pkg/redis"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := logger.New(cfg.LogLevel)

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Connect to Redis
	redisClient := redis.NewClient(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)

	// Setup Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(r, db, redisClient, logger)

	// Start server
	logger.Info("Server starting on port " + cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
