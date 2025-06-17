package services

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/waltertaya/server_check_bd/internal/logger"
	"github.com/waltertaya/server_check_bd/internal/models"
)

// HealthChecker represents a service that checks server health
type HealthChecker struct {
	serverService *ServerService
	clients       map[int]chan models.ServerStatus
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(serverService *ServerService) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		serverService: serverService,
		clients:       make(map[int]chan models.ServerStatus),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the health checking process
func (hc *HealthChecker) Start() {
	logger.Info("Starting health checker")
	go hc.checkLoop()
}

// Stop stops the health checking process
func (hc *HealthChecker) Stop() {
	logger.Info("Stopping health checker")
	hc.cancel()
}

// Subscribe adds a new client to receive status updates
func (hc *HealthChecker) Subscribe(serverID int) chan models.ServerStatus {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	ch := make(chan models.ServerStatus, 1)
	hc.clients[serverID] = ch
	return ch
}

// Unsubscribe removes a client from receiving status updates
func (hc *HealthChecker) Unsubscribe(serverID int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if ch, exists := hc.clients[serverID]; exists {
		close(ch)
		delete(hc.clients, serverID)
	}
}

// checkLoop continuously checks server health
func (hc *HealthChecker) checkLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			servers, err := hc.serverService.GetServers()
			if err != nil {
				logger.Error("Failed to get servers: %v", err)
				continue
			}

			for _, server := range servers {
				go hc.checkServer(server)
			}
		}
	}
}

// checkServer checks the health of a single server
func (hc *HealthChecker) checkServer(server models.Server) {
	start := time.Now()
	client := &http.Client{
		Timeout: time.Duration(server.Timeout) * time.Millisecond,
	}

	req, err := http.NewRequest(server.Method, server.URL, nil)
	if err != nil {
		hc.handleError(server, err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		hc.handleError(server, err)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	status := models.ServerStatus{
		IsUp:         resp.StatusCode == server.ExpectedStatus,
		StatusCode:   &resp.StatusCode,
		ResponseTime: intPtr(int(duration.Milliseconds())),
		LastChecked:  time.Now(),
	}

	// Update server status
	err = hc.serverService.UpdateServerStatus(server.ID, status)
	if err != nil {
		logger.Error("Failed to update server status: %v", err)
		return
	}

	// Notify subscribers
	hc.mu.RLock()
	if ch, exists := hc.clients[server.ID]; exists {
		select {
		case ch <- status:
		default:
			// Channel is full, skip this update
		}
	}
	hc.mu.RUnlock()
}

// handleError handles errors during server health checks
func (hc *HealthChecker) handleError(server models.Server, err error) {
	errorMsg := err.Error()
	status := models.ServerStatus{
		IsUp:        false,
		Error:       &errorMsg,
		LastChecked: time.Now(),
	}

	err = hc.serverService.UpdateServerStatus(server.ID, status)
	if err != nil {
		logger.Error("Failed to update server status: %v", err)
		return
	}

	// Notify subscribers
	hc.mu.RLock()
	if ch, exists := hc.clients[server.ID]; exists {
		select {
		case ch <- status:
		default:
			// Channel is full, skip this update
		}
	}
	hc.mu.RUnlock()
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}
