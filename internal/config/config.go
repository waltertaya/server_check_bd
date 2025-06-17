package config

import (
	"os"
	"path/filepath"
	"time"
)

var (
	// DataDir is the directory where data files are stored
	DataDir string

	// ServersFile is the path to the servers JSON file
	ServersFile string

	// HistoryFile is the path to the history JSON file
	HistoryFile string

	// LogDir is the directory where log files are stored
	LogDir string

	// DefaultCheckInterval is the default interval for checking server health
	DefaultCheckInterval = 60 * time.Second

	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 5 * time.Second
)

// Init initializes the configuration
func Init() {
	// Set up data directory
	DataDir = getEnv("DATA_DIR", "data")
	ServersFile = filepath.Join(DataDir, "servers.json")
	HistoryFile = filepath.Join(DataDir, "history.json")

	// Set up logs directory
	LogDir = getEnv("LOG_DIR", "logs")

	// Create directories if they don't exist
	os.MkdirAll(DataDir, 0755)
	os.MkdirAll(LogDir, 0755)
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}