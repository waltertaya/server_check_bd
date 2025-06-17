package services

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/waltertaya/server_check_bd/internal/logger"
	"github.com/waltertaya/server_check_bd/internal/models"
)

// ServerService handles server-related operations
type ServerService struct {
	db *sqlx.DB
}

// NewServerService creates a new server service instance
func NewServerService(db *sqlx.DB) *ServerService {
	return &ServerService{
		db: db,
	}
}

// CreateServer creates a new server
func (s *ServerService) CreateServer(req models.CreateServerRequest) (*models.Server, error) {
	server := &models.Server{
		Name:           req.Name,
		Description:    "",
		URL:            req.URL,
		Method:         req.Method,
		Interval:       req.Interval,
		Timeout:        req.Timeout,
		ExpectedStatus: req.ExpectedStatus,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if req.Description != nil {
		server.Description = *req.Description
	}

	result, err := s.db.NamedExec(`
		INSERT INTO servers (name, description, url, method, interval, timeout, expected_status, created_at, updated_at)
		VALUES (:name, :description, :url, :method, :interval, :timeout, :expected_status, :created_at, :updated_at)
	`, server)
	if err != nil {
		logger.Error("Failed to create server: %v", err)
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("Failed to get last insert ID: %v", err)
		return nil, err
	}

	server.ID = int(id)
	return server, nil
}

// GetServers returns all servers
func (s *ServerService) GetServers() ([]models.Server, error) {
	var servers []models.Server
	err := s.db.Select(&servers, "SELECT * FROM servers ORDER BY created_at DESC")
	if err != nil {
		logger.Error("Failed to get servers: %v", err)
		return nil, err
	}
	return servers, nil
}

// GetServerByID returns a server by its ID
func (s *ServerService) GetServerByID(id int) (*models.Server, error) {
	var server models.Server
	err := s.db.Get(&server, "SELECT * FROM servers WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.Error("Failed to get server %d: %v", id, err)
		return nil, err
	}
	return &server, nil
}

// UpdateServer updates a server
func (s *ServerService) UpdateServer(id int, req models.UpdateServerRequest) (*models.Server, error) {
	server, err := s.GetServerByID(id)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, nil
	}

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

	server.UpdatedAt = time.Now()

	_, err = s.db.NamedExec(`
		UPDATE servers
		SET name = :name,
			url = :url,
			method = :method,
			expected_status = :expected_status,
			timeout = :timeout,
			interval = :interval,
			updated_at = :updated_at
		WHERE id = :id
	`, server)
	if err != nil {
		logger.Error("Failed to update server %d: %v", id, err)
		return nil, err
	}

	return server, nil
}

// DeleteServer deletes a server
func (s *ServerService) DeleteServer(id int) error {
	_, err := s.db.Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		logger.Error("Failed to delete server %d: %v", id, err)
		return err
	}
	return nil
}

// UpdateServerStatus updates a server's status
func (s *ServerService) UpdateServerStatus(id int, status models.ServerStatus) error {
	history := models.ServerHistory{
		ServerID:     id,
		IsUp:         status.IsUp,
		StatusCode:   status.StatusCode,
		ResponseTime: status.ResponseTime,
		ResponseBody: status.ResponseBody,
		Error:        status.Error,
		CheckedAt:    status.LastChecked,
	}

	_, err := s.db.NamedExec(`
		INSERT INTO status_history (server_id, is_up, status_code, response_time, response_body, error, checked_at)
		VALUES (:server_id, :is_up, :status_code, :response_time, :response_body, :error, :checked_at)
	`, history)
	if err != nil {
		logger.Error("Failed to insert status history for server %d: %v", id, err)
		return err
	}

	return nil
}

// GetServerHistory returns the status history for a server
func (s *ServerService) GetServerHistory(id int, limit int) ([]models.ServerHistory, error) {
	var history []models.ServerHistory
	err := s.db.Select(&history, `
		SELECT * FROM status_history
		WHERE server_id = ?
		ORDER BY checked_at DESC
		LIMIT ?
	`, id, limit)
	if err != nil {
		logger.Error("Failed to get server history for server %d: %v", id, err)
		return nil, err
	}
	return history, nil
}

// GetLatestStatus returns the latest status for a server
func (s *ServerService) GetLatestStatus(id int) (*models.ServerStatus, error) {
	var history models.ServerHistory
	err := s.db.Get(&history, `
		SELECT * FROM status_history
		WHERE server_id = ?
		ORDER BY checked_at DESC
		LIMIT 1
	`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.Error("Failed to get latest status for server %d: %v", id, err)
		return nil, err
	}

	status := &models.ServerStatus{
		IsUp:         history.IsUp,
		StatusCode:   history.StatusCode,
		ResponseTime: history.ResponseTime,
		ResponseBody: history.ResponseBody,
		Error:        history.Error,
		LastChecked:  history.CheckedAt,
	}

	return status, nil
}
