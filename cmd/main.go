package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/waltertaya/server_check_bd/internal/db"
	"github.com/waltertaya/server_check_bd/internal/handlers"
	"github.com/waltertaya/server_check_bd/internal/logger"
	"github.com/waltertaya/server_check_bd/internal/services"
)

func main() {
	// Initialize logger
	if err := logger.Init(logger.INFO, "logs/app.log"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize database
	database, err := db.ConnectDB()
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		return
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		logger.Error("Failed to run migrations: %v", err)
		return
	}

	// Initialize services
	serverService := services.NewServerService(database)
	healthChecker := services.NewHealthChecker(serverService)
	go healthChecker.Start()

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(database)
	serverHandlers := handlers.NewServerHandlers(serverService, healthChecker)
	wsHandler := handlers.NewWebSocketHandler(healthChecker)

	// Initialize router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Auth routes
	router.POST("/api/auth/register", authHandlers.RegisterHandler)
	router.POST("/api/auth/login", authHandlers.LoginHandler)

	// Server routes
	router.GET("/api/servers", serverHandlers.GetServers)
	router.GET("/api/servers/:id", serverHandlers.GetServer)
	router.POST("/api/servers", serverHandlers.CreateServer)
	router.PUT("/api/servers/:id", serverHandlers.UpdateServer)
	router.DELETE("/api/servers/:id", serverHandlers.DeleteServer)
	router.GET("/api/servers/:id/history", serverHandlers.GetServerHistory)

	// WebSocket route
	router.GET("/api/servers/:id/ws", wsHandler.HandleWebSocket)

	// Start server
	if err := router.Run(":8080"); err != nil {
		logger.Error("Failed to start server: %v", err)
	}
}
