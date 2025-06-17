package models

import "time"

// Server represents a server to be monitored
type Server struct {
	ID             int       `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Description    string    `db:"description" json:"description"`
	URL            string    `db:"url" json:"url"`
	Method         string    `db:"method" json:"method"`
	Interval       int       `db:"interval" json:"interval"`
	Timeout        int       `db:"timeout" json:"timeout"`
	ExpectedStatus int       `db:"expected_status" json:"expectedStatus"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at" json:"updatedAt"`
}

// ServerStatus represents the current status of a server
type ServerStatus struct {
	IsUp         bool      `db:"is_up" json:"isUp"`
	StatusCode   *int      `db:"status_code" json:"statusCode"`
	ResponseTime *int      `db:"response_time" json:"responseTime"`
	ResponseBody *string   `db:"response_body" json:"responseBody"`
	Error        *string   `db:"error" json:"error"`
	LastChecked  time.Time `db:"checked_at" json:"lastChecked"`
	State        string    `json:"state"`
}

// ServerHistory represents a historical status record
type ServerHistory struct {
	ID           int       `db:"id" json:"id"`
	ServerID     int       `db:"server_id" json:"serverId"`
	IsUp         bool      `db:"is_up" json:"isUp"`
	StatusCode   *int      `db:"status_code" json:"statusCode"`
	ResponseTime *int      `db:"response_time" json:"responseTime"`
	ResponseBody *string   `db:"response_body" json:"responseBody"`
	Error        *string   `db:"error" json:"error"`
	CheckedAt    time.Time `db:"checked_at" json:"checkedAt"`
	State        string    `json:"state"`
}

// CreateServerRequest represents the request to create a new server
type CreateServerRequest struct {
	Name           string  `json:"name" binding:"required"`
	URL            string  `json:"url" binding:"required,url"`
	Description    *string `json:"description,omitempty"`
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
