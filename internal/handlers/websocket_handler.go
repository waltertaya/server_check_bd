package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/waltertaya/server_check_bd/internal/logger"
	"github.com/waltertaya/server_check_bd/internal/services"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	checker *services.HealthChecker
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(checker *services.HealthChecker) *WebSocketHandler {
	return &WebSocketHandler{
		checker: checker,
	}
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	serverID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid server ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Subscribe to server status updates
	statusChan := h.checker.Subscribe(serverID)
	defer h.checker.Unsubscribe(serverID)

	// Handle incoming messages
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				logger.Error("Error reading message: %v", err)
				return
			}
		}
	}()

	// Send status updates
	for status := range statusChan {
		err := conn.WriteJSON(map[string]interface{}{
			"type":   "server:status",
			"id":     serverID,
			"status": status,
		})
		if err != nil {
			logger.Error("Error writing message: %v", err)
			return
		}
	}
}
