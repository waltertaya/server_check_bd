package services

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"pulse/config"
	"pulse/models"
)

// ServerService handles server CRUD operations
type ServerService struct {
	mutex     sync.RWMutex
	servers   []models.Server
	histories []models.StatusHistory
}

// NewServerService creates a new server service
func NewServerService() *ServerService {
	service := &ServerService{}
	service.loadServers()
	service.loadHistories()
	return service
}

// GetServers returns all servers
func (s *ServerService) GetServers() []models.Server {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.servers
}

// GetServerByID returns a server by ID
func (s *ServerService) GetServerByID(id string) (models.Server, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, server := range s.servers {
		if server.ID == id {
			return server, nil
		}
	}

	return models.Server{}, errors.New("server not found")
}

// CreateServer creates a new server
func (s *ServerService) CreateServer(req models.CreateServerRequest) (models.Server, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	id := generateID()

	server := models.Server{
		ID:             id,
		Name:           req.Name,
		Description:    req.Description,
		URL:            req.URL,
		Method:         req.Method,
		ExpectedStatus: req.ExpectedStatus,
		Timeout:        req.Timeout,
		Interval:       req.Interval,
		CreatedAt:      now,
	}

	s.servers = append(s.servers, server)
	err := s.saveServers()
	if err != nil {
		return models.Server{}, err
	}

	return server, nil
}

// UpdateServer updates an existing server
func (s *ServerService) UpdateServer(id string, req models.UpdateServerRequest) (models.Server, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, server := range s.servers {
		if server.ID == id {
			if req.Name != nil {
				server.Name = *req.Name
			}
			if req.URL != nil {
				server.URL = *req.URL
			}
			if req.Method != nil {
				server.Method = *req.Method
			}
			if req.ExpectedStatus != nil {
				server.ExpectedStatus = *req.ExpectedStatus
			}
			if req.Timeout != nil {
				server.Timeout = *req.Timeout
			}
			if req.Interval != nil {
				server.Interval = *req.Interval
			}

			now := time.Now()
			server.UpdatedAt = &now

			s.servers[i] = server
			err := s.saveServers()
			if err != nil {
				return models.Server{}, err
			}

			return server, nil
		}
	}

	return models.Server{}, errors.New("server not found")
}

// DeleteServer deletes a server
func (s *ServerService) DeleteServer(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, server := range s.servers {
		if server.ID == id {
			s.servers = append(s.servers[:i], s.servers[i+1:]...)
			return s.saveServers()
		}
	}

	return errors.New("server not found")
}

// UpdateServerStatus updates the status of a server
func (s *ServerService) UpdateServerStatus(id string, status models.ServerStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, server := range s.servers {
		if server.ID == id {
			server.LastStatus = &status
			s.servers[i] = server

			// Save status to history
			history := models.StatusHistory{
				ServerID:     id,
				Timestamp:    status.LastChecked,
				IsUp:         status.IsUp,
				StatusCode:   status.StatusCode,
				ResponseTime: status.ResponseTime,
				Error:        status.Error,
				State:        status.State,
			}
			s.histories = append(s.histories, history)

			// Limit history size
			if len(s.histories) > 1000 {
				s.histories = s.histories[len(s.histories)-1000:]
			}

			err := s.saveServers()
			if err != nil {
				return err
			}

			return s.saveHistories()
		}
	}

	return errors.New("server not found")
}

// GetServerHistory returns the history for a server
func (s *ServerService) GetServerHistory(id string, limit int) ([]models.StatusHistory, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var serverHistory []models.StatusHistory
	for _, history := range s.histories {
		if history.ServerID == id {
			serverHistory = append(serverHistory, history)
		}
	}

	// Check if there's any history
	if len(serverHistory) == 0 {
		return []models.StatusHistory{}, nil
	}

	// Limit the results
	if limit > 0 && len(serverHistory) > limit {
		start := len(serverHistory) - limit
		if start < 0 {
			start = 0
		}
		serverHistory = serverHistory[start:]
	}

	return serverHistory, nil
}

// Private methods

func (s *ServerService) loadServers() error {
	// Check if file exists
	if _, err := os.Stat(config.ServersFile); os.IsNotExist(err) {
		// Create default servers if file doesn't exist
		s.servers = getDefaultServers()
		return s.saveServers()
	}

	// Read file
	data, err := ioutil.ReadFile(config.ServersFile)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	err = json.Unmarshal(data, &s.servers)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerService) saveServers() error {
	data, err := json.MarshalIndent(s.servers, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(config.ServersFile, data, 0644)
}

func (s *ServerService) loadHistories() error {
	// Check if file exists
	if _, err := os.Stat(config.HistoryFile); os.IsNotExist(err) {
		// Create empty history if file doesn't exist
		s.histories = []models.StatusHistory{}
		return s.saveHistories()
	}

	// Read file
	data, err := ioutil.ReadFile(config.HistoryFile)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	err = json.Unmarshal(data, &s.histories)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerService) saveHistories() error {
	data, err := json.MarshalIndent(s.histories, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(config.HistoryFile, data, 0644)
}

// Helper functions

func generateID() string {
	return time.Now().Format("20060102150405.000")
}

func getDefaultServers() []models.Server {
	now := time.Now()
	return []models.Server{
		{
			ID:             "1",
			Name:           "Google",
			URL:            "https://www.google.com",
			Method:         "GET",
			ExpectedStatus: 200,
			Timeout:        5000,
			Interval:       60000,
			CreatedAt:      now,
		},
		{
			ID:             "2",
			Name:           "Example API",
			URL:            "https://jsonplaceholder.typicode.com/posts/1",
			Method:         "GET",
			ExpectedStatus: 200,
			Timeout:        5000,
			Interval:       120000,
			CreatedAt:      now,
		},
	}
}