package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/waltertaya/server_check_bd/internal/logger"
	"github.com/waltertaya/server_check_bd/internal/models"
	"github.com/waltertaya/server_check_bd/internal/services"
)

// ServerHandlers handles server-related HTTP requests
type ServerHandlers struct {
	service *services.ServerService
	checker *services.HealthChecker
}

// NewServerHandlers creates a new server handlers instance
func NewServerHandlers(service *services.ServerService, checker *services.HealthChecker) *ServerHandlers {
	return &ServerHandlers{
		service: service,
		checker: checker,
	}
}

// GetServer handles GET /api/servers/:id
func (h *ServerHandlers) GetServer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid server ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	server, err := h.service.GetServerByID(id)
	if err != nil {
		logger.Error("Failed to get server: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get server"})
		return
	}

	if server == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// CreateServer handles POST /api/servers
func (h *ServerHandlers) CreateServer(c *gin.Context) {
	var req models.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	server, err := h.service.CreateServer(req)
	if err != nil {
		logger.Error("Failed to create server: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, server)
}

// UpdateServer handles PUT /api/servers/:id
func (h *ServerHandlers) UpdateServer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid server ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req models.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	server, err := h.service.UpdateServer(id, req)
	if err != nil {
		logger.Error("Failed to update server: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	if server == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// DeleteServer handles DELETE /api/servers/:id
func (h *ServerHandlers) DeleteServer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid server ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	err = h.service.DeleteServer(id)
	if err != nil {
		logger.Error("Failed to delete server: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetServers handles GET /api/servers
func (h *ServerHandlers) GetServers(c *gin.Context) {
	servers, err := h.service.GetServers()
	if err != nil {
		logger.Error("Failed to get servers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get servers"})
		return
	}

	c.JSON(http.StatusOK, servers)
}

// GetServerHistory handles GET /api/servers/:id/history
func (h *ServerHandlers) GetServerHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid server ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	limit := 100 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	history, err := h.service.GetServerHistory(id, limit)
	if err != nil {
		logger.Error("Failed to get server history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get server history"})
		return
	}

	c.JSON(http.StatusOK, history)
}
