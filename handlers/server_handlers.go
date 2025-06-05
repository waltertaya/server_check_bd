package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pulse/models"
	"pulse/services"
)

// GetServers returns all servers
func GetServers(service *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		servers := service.GetServers()
		c.JSON(http.StatusOK, servers)
	}
}

// GetServer returns a server by ID
func GetServer(service *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		server, err := service.GetServerByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
			return
		}
		c.JSON(http.StatusOK, server)
	}
}

// CreateServer creates a new server
func CreateServer(service *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateServerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		server, err := service.CreateServer(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, server)
	}
}

// UpdateServer updates a server
func UpdateServer(service *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req models.UpdateServerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		server, err := service.UpdateServer(id, req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, server)
	}
}

// DeleteServer deletes a server
func DeleteServer(service *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		err := service.DeleteServer(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// CheckServer checks the health of a server
func CheckServer(checker *services.HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		status, err := checker.CheckServer(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

// GetServerHistory returns the history for a server
func GetServerHistory(service *services.ServerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		// Parse limit query parameter
		limit := 50
		limitStr := c.DefaultQuery("limit", "50")
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}

		history, err := service.GetServerHistory(id, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, history)
	}
}