package db

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// ConnectDB establishes a connection to the SQLite database and returns the database instance
func ConnectDB() (*sqlx.DB, error) {
	// Ensure the data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open SQLite database
	db, err := sqlx.Connect("sqlite3", "data/server_monitor.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations executes database migrations
func RunMigrations(db *sqlx.DB) error {
	// Drop existing tables if they exist
	_, err := db.Exec(`
		DROP TABLE IF EXISTS status_history;
		DROP TABLE IF EXISTS servers;
		DROP TABLE IF EXISTS users;
	`)
	if err != nil {
		return fmt.Errorf("failed to drop existing tables: %v", err)
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create servers table
	_, err = db.Exec(`
		CREATE TABLE servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			url TEXT NOT NULL,
			method TEXT NOT NULL,
			interval INTEGER NOT NULL,
			timeout INTEGER NOT NULL,
			expected_status INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create servers table: %v", err)
	}

	// Create status_history table
	_, err = db.Exec(`
		CREATE TABLE status_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER NOT NULL,
			is_up BOOLEAN NOT NULL,
			status_code INTEGER,
			response_time INTEGER,
			response_body TEXT,
			error TEXT,
			checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (server_id) REFERENCES servers(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create status_history table: %v", err)
	}

	return nil
}
