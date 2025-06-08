package services

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/waltertaya/server_check_bd/models"
)

// HealthChecker is responsible for checking server health
type HealthChecker struct {
	serverService     *ServerService
	clients           map[*Client]bool
	clientsMutex      sync.RWMutex
	statusUpdateChan  chan statusUpdate
	checkRequestChan  chan string
	serverCheckTimers map[string]*time.Timer
	serverTimersMutex sync.RWMutex
}

// Client represents a connected WebSocket client
type Client struct {
	Send chan interface{}
}

// statusUpdate represents a status update for a server
type statusUpdate struct {
	ServerID string
	Status   models.ServerStatus
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(serverService *ServerService) *HealthChecker {
	return &HealthChecker{
		serverService:     serverService,
		clients:           make(map[*Client]bool),
		statusUpdateChan:  make(chan statusUpdate, 100),
		checkRequestChan:  make(chan string, 100),
		serverCheckTimers: make(map[string]*time.Timer),
	}
}

// Start begins the health checker service
func (hc *HealthChecker) Start() {
	// Start status broadcaster
	go hc.broadcastStatusUpdates()

	// Initialize timers for all servers
	servers, _ := hc.serverService.GetServers()
	for _, server := range servers {
		hc.scheduleServerCheck(server)
	}

	// Listen for check requests
	go hc.handleCheckRequests()
}

// CheckServer checks the health of a specific server
func (hc *HealthChecker) CheckServer(id string) (models.ServerStatus, error) {
	server, err := hc.serverService.GetServerByID(id)
	if err != nil {
		return models.ServerStatus{}, err
	}

	status := hc.checkServerHealth(server)

	// Update server status
	err = hc.serverService.UpdateServerStatus(id, status)
	if err != nil {
		return models.ServerStatus{}, err
	}

	// Send status update to all clients
	hc.statusUpdateChan <- statusUpdate{
		ServerID: id,
		Status:   status,
	}

	return status, nil
}

// RequestCheckServer requests a check for a specific server
func (hc *HealthChecker) RequestCheckServer(id string) {
	hc.checkRequestChan <- id
}

// RegisterClient registers a new WebSocket client
func (hc *HealthChecker) RegisterClient(client *Client) {
	hc.clientsMutex.Lock()
	defer hc.clientsMutex.Unlock()
	hc.clients[client] = true
}

// UnregisterClient unregisters a WebSocket client
func (hc *HealthChecker) UnregisterClient(client *Client) {
	hc.clientsMutex.Lock()
	defer hc.clientsMutex.Unlock()
	if _, ok := hc.clients[client]; ok {
		delete(hc.clients, client)
		close(client.Send)
	}
}

// Private methods

func (hc *HealthChecker) handleCheckRequests() {
	for id := range hc.checkRequestChan {
		hc.CheckServer(id)
	}
}

func (hc *HealthChecker) broadcastStatusUpdates() {
	for update := range hc.statusUpdateChan {
		hc.clientsMutex.RLock()
		for client := range hc.clients {
			select {
			case client.Send <- map[string]interface{}{
				"type":   "server:status",
				"id":     update.ServerID,
				"status": update.Status,
			}:
			default:
				close(client.Send)
				delete(hc.clients, client)
			}
		}
		hc.clientsMutex.RUnlock()
	}
}

func (hc *HealthChecker) scheduleServerCheck(server models.Server) {
	hc.serverTimersMutex.Lock()
	defer hc.serverTimersMutex.Unlock()

	// Cancel existing timer if any
	if timer, exists := hc.serverCheckTimers[server.ID]; exists {
		timer.Stop()
	}

	// Calculate interval in milliseconds
	interval := time.Duration(server.Interval) * time.Millisecond

	// Create a new timer
	hc.serverCheckTimers[server.ID] = time.AfterFunc(interval, func() {
		// Check server health
		status := hc.checkServerHealth(server)

		// Update server status
		err := hc.serverService.UpdateServerStatus(server.ID, status)
		if err != nil {
			fmt.Printf("Error updating server status: %v\n", err)
		}

		// Send status update to all clients
		hc.statusUpdateChan <- statusUpdate{
			ServerID: server.ID,
			Status:   status,
		}

		// Reschedule next check
		hc.scheduleServerCheck(server)
	})
}

func (hc *HealthChecker) checkServerHealth(server models.Server) models.ServerStatus {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(server.Timeout) * time.Millisecond,
	}

	// Create request
	req, err := http.NewRequest(server.Method, server.URL, nil)
	if err != nil {
		errorStr := err.Error()
		return models.ServerStatus{
			IsUp:        false,
			LastChecked: time.Now(),
			Error:       &errorStr,
			State:       "down",
		}
	}

	// Set common headers
	req.Header.Set("User-Agent", "Pulse-Server-Monitor/1.0")

	// Record start time
	startTime := time.Now()

	// Send request
	resp, err := client.Do(req)

	// Calculate response time
	responseTime := int(time.Since(startTime).Milliseconds())

	// Handle error
	if err != nil {
		errorStr := err.Error()
		return models.ServerStatus{
			IsUp:        false,
			LastChecked: time.Now(),
			Error:       &errorStr,
			State:       "down",
		}
	}
	defer resp.Body.Close()

	// Check status code
	statusCode := resp.StatusCode
	isUp := statusCode == server.ExpectedStatus

	// Determine state
	state := "down"
	if isUp {
		state = "healthy"
		// If response time is high, mark as warning
		if responseTime > server.Timeout/2 {
			state = "warning"
		}
	} else if statusCode >= 200 && statusCode < 400 {
		// If status code is different but still successful, mark as warning
		state = "warning"
	}

	return models.ServerStatus{
		IsUp:         isUp,
		StatusCode:   &statusCode,
		ResponseTime: &responseTime,
		LastChecked:  time.Now(),
		State:        state,
	}
}
