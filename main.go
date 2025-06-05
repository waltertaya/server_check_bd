package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/net/websocket"

	"pulse/config"
	"pulse/handlers"
	"pulse/services"
)

func main() {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load()
	config.Init()

	// Initialize services
	serverService := services.NewServerService()
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

	// 1) Register all /api routes
	api := router.Group("/api")
	{
		api.GET("/servers", handlers.GetServers(serverService))
		api.GET("/servers/:id", handlers.GetServer(serverService))
		api.POST("/servers", handlers.CreateServer(serverService))
		api.PUT("/servers/:id", handlers.UpdateServer(serverService))
		api.DELETE("/servers/:id", handlers.DeleteServer(serverService))
		api.POST("/servers/:id/check", handlers.CheckServer(healthChecker))
		api.GET("/servers/:id/history", handlers.GetServerHistory(serverService))
	}

	// 2) Serve static assets from the Vite build
	//    Vite’s dist/ output typically has:
	//      - index.html
	//      - assets/ (JS/CSS/images)
	//      - favicon.ico, etc.
	//
	//    So we expose "/assets/*" and any known files, then use NoRoute
	//    to serve index.html on all other non-API requests (SPA fallback).

	// Serve the "assets" directory
	router.Static("/assets", "./frontend/dist/assets")

	// Serve other static files if they exist (e.g. favicon, robots.txt)
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	router.StaticFile("/robots.txt", "./frontend/dist/robots.txt")

	// 3) SPA fallback: for any GET that didn’t match /api or /assets, serve index.html
	router.NoRoute(func(c *gin.Context) {
		// Only intercept GET requests (so API POST/PUT/etc. still return 404 if not found)
		if c.Request.Method == http.MethodGet {
			c.File("./frontend/dist/index.html")
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	// 4) WebSocket handler on a separate net/http mux
	wsServer := websocket.Server{
		Handler: handlers.WebSocketHandler(healthChecker),
	}
	http.Handle("/ws", wsServer)

	// 5) Start listening
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(router.Run(":" + port))
}
