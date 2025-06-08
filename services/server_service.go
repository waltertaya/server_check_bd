package services

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/waltertaya/server_check_bd/models"
)

// ServerService handles server CRUD operations backed by SQLite DB via sqlx
type ServerService struct {
	db *sqlx.DB
}

// NewServerService creates a new service with given DB connection
func NewServerService(db *sqlx.DB) *ServerService {
	return &ServerService{db: db}
}

// GetServers returns all servers
func (s *ServerService) GetServers() ([]models.Server, error) {
	var servers []models.Server
	err := s.db.Select(&servers, `
		SELECT id, name, description, url, method, expected_status AS expectedStatus,
		       timeout, interval, created_at AS createdAt, updated_at AS updatedAt
		FROM servers
	`)
	return servers, err
}

// GetServerByID returns a server by ID
func (s *ServerService) GetServerByID(id string) (models.Server, error) {
	var server models.Server
	err := s.db.Get(&server, `
		SELECT id, name, description, url, method, expected_status AS expectedStatus,
		       timeout, interval, created_at AS createdAt, updated_at AS updatedAt, last_status AS lastStatus
		FROM servers WHERE id = ?
	`, id)
	if err != nil {
		return models.Server{}, err
	}
	return server, nil
}

// CreateServer creates a new server
func (s *ServerService) CreateServer(req models.CreateServerRequest) (models.Server, error) {
	now := time.Now()
	res, err := s.db.Exec(`
		INSERT INTO servers (name, description, url, method, expected_status, timeout, interval, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, req.Name, req.Description, req.URL, req.Method, req.ExpectedStatus,
		req.Timeout, req.Interval, now)
	if err != nil {
		return models.Server{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Server{}, err
	}
	return s.GetServerByID(fmt.Sprint(id))
}

// UpdateServer updates an existing server
func (s *ServerService) UpdateServer(id string, req models.UpdateServerRequest) (models.Server, error) {
	// Build dynamic query parts
	query := "UPDATE servers SET"
	args := []interface{}{}
	if req.Name != nil {
		query += " name = ?,"
		args = append(args, *req.Name)
	}
	if req.URL != nil {
		query += " url = ?,"
		args = append(args, *req.URL)
	}
	if req.Method != nil {
		query += " method = ?,"
		args = append(args, *req.Method)
	}
	if req.ExpectedStatus != nil {
		query += " expected_status = ?,"
		args = append(args, *req.ExpectedStatus)
	}
	if req.Timeout != nil {
		query += " timeout = ?,"
		args = append(args, *req.Timeout)
	}
	if req.Interval != nil {
		query += " interval = ?,"
		args = append(args, *req.Interval)
	}
	// set updated_at
	now := time.Now()
	query += " updated_at = ?"
	args = append(args, now)

	// finalize
	query += " WHERE id = ?"
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return models.Server{}, err
	}
	return s.GetServerByID(id)
}

// DeleteServer deletes a server
func (s *ServerService) DeleteServer(id string) error {
	_, err := s.db.Exec("DELETE FROM servers WHERE id = ?", id)
	return err
}

// UpdateServerStatus updates the status and records history
func (s *ServerService) UpdateServerStatus(id string, status models.ServerStatus) error {
	// update last_status in servers
	_, err := s.db.NamedExec(`
		UPDATE servers SET last_status = :lastStatus WHERE id = :id
	`, map[string]interface{}{"id": id, "lastStatus": status})
	if err != nil {
		return err
	}
	// insert into history
	hist := models.StatusHistory{
		ServerID:     id,
		Timestamp:    status.LastChecked,
		IsUp:         status.IsUp,
		StatusCode:   status.StatusCode,
		ResponseTime: status.ResponseTime,
		Error:        status.Error,
		State:        status.State,
	}
	_, err = s.db.NamedExec(`
		INSERT INTO status_history (server_id, timestamp, is_up, status_code, response_time, error, state)
		VALUES (:server_id, :timestamp, :is_up, :status_code, :response_time, :error, :state)
	`, hist)
	return err
}

// GetServerHistory returns the history for a server
func (s *ServerService) GetServerHistory(id string, limit int) ([]models.StatusHistory, error) {
	query := `SELECT server_id AS serverId, timestamp, is_up AS isUp, status_code AS statusCode,
		response_time AS responseTime, error, state FROM status_history WHERE server_id = ? ORDER BY timestamp DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	var histories []models.StatusHistory
	err := s.db.Select(&histories, query, id)
	// reverse to ascending
	for i, j := 0, len(histories)-1; i < j; i, j = i+1, j-1 {
		histories[i], histories[j] = histories[j], histories[i]
	}
	return histories, err
}
