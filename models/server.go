package models

import "time"

type User struct {
	ID       string `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Email    string `db:"email" json:"email"`
	Password string `db:"password,omitempty"`
}

// Server represents a server to be monitored
type Server struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    *string       `json:"description,omitempty"` // Optional description of the server
	URL            string        `json:"url"`
	Method         string        `json:"method"`
	ExpectedStatus int           `json:"expectedStatus"`
	Timeout        int           `json:"timeout"`  // in milliseconds
	Interval       int           `json:"interval"` // in milliseconds
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      *time.Time    `json:"updatedAt,omitempty"`
	LastStatus     *ServerStatus `json:"lastStatus,omitempty"`
}

// ServerStatus represents the status of a server check
type ServerStatus struct {
	IsUp         bool      `json:"isUp"`
	StatusCode   *int      `json:"statusCode,omitempty"`
	ResponseTime *int      `json:"responseTime,omitempty"` // in milliseconds
	LastChecked  time.Time `json:"lastChecked"`
	Error        *string   `json:"error,omitempty"`
	State        string    `json:"state"` // "healthy", "warning", "down", "unknown"
}

// StatusHistory represents a historical status check
type StatusHistory struct {
	ServerID     string    `json:"serverId"`
	Timestamp    time.Time `json:"timestamp"`
	IsUp         bool      `json:"isUp"`
	StatusCode   *int      `json:"statusCode,omitempty"`
	ResponseTime *int      `json:"responseTime,omitempty"`
	Error        *string   `json:"error,omitempty"`
	State        string    `json:"state"`
}

// CreateServerRequest represents the request to create a new server
type CreateServerRequest struct {
	Name           string  `json:"name" binding:"required"`
	URL            string  `json:"url" binding:"required,url"`
	Description    *string `json:"description,omitempty"` // Optional description of the server
	Method         string  `json:"method" binding:"required,oneof=GET POST HEAD"`
	ExpectedStatus int     `json:"expectedStatus" binding:"required,min=100,max=599"`
	Timeout        int     `json:"timeout" binding:"required,min=1000"`
	Interval       int     `json:"interval" binding:"required,min=5000"`
}

// UpdateServerRequest represents the request to update a server
type UpdateServerRequest struct {
	Name           *string `json:"name"`
	URL            *string `json:"url" binding:"omitempty,url"`
	Method         *string `json:"method" binding:"omitempty,oneof=GET POST HEAD"`
	ExpectedStatus *int    `json:"expectedStatus" binding:"omitempty,min=100,max=599"`
	Timeout        *int    `json:"timeout" binding:"omitempty,min=1000"`
	Interval       *int    `json:"interval" binding:"omitempty,min=5000"`
}
