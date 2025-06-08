package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/waltertaya/server_check_bd/db"
	"github.com/waltertaya/server_check_bd/handlers"
	"github.com/waltertaya/server_check_bd/initialisers"
	"github.com/waltertaya/server_check_bd/services"
	"golang.org/x/net/websocket"
)

func init() {
	initialisers.LoadEnv()
}

func main() {

	err := db.Connect()
	if err == nil {
		fmt.Println("Connected to database successfully")
	}

	// Initialize services with DB
	serverService := services.NewServerService(db.DB)
	healthChecker := services.NewHealthChecker(serverService)
	go healthChecker.Start()

	// Initialize Gin router
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

	// Register all /api routes
	api := router.Group("/api")
	{
		// Authentication routes
		api.POST("/auth/login", handlers.LoginHandler)
		api.POST("/auth/register", handlers.RegisterHandler)
		// Server management routes
		api.GET("/servers", handlers.GetServers(serverService))
		api.GET("/servers/:id", handlers.GetServer(serverService))
		api.POST("/servers", handlers.CreateServer(serverService))
		api.PUT("/servers/:id", handlers.UpdateServer(serverService))
		api.DELETE("/servers/:id", handlers.DeleteServer(serverService))
		api.POST("/servers/:id/check", handlers.CheckServer(healthChecker))
		api.GET("/servers/:id/history", handlers.GetServerHistory(serverService))
	}

	// WebSocket handler on a separate net/http mux
	wsServer := websocket.Server{
		Handler: handlers.WebSocketHandler(healthChecker),
	}
	http.Handle("/ws", wsServer)

	// Start listening
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(router.Run(":" + port))
}
