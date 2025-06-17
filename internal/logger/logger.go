package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents the severity of the log
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger is the main logger struct
type Logger struct {
	*log.Logger
	level LogLevel
}

var (
	// Default logger instance
	defaultLogger *Logger
)

// Init initializes the logger with the given log level and output file
func Init(level LogLevel, logFile string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Initialize default logger
	defaultLogger = &Logger{
		Logger: log.New(multiWriter, "", 0),
		level:  level,
	}

	return nil
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3) // Skip 3 levels to get the actual caller
	if !ok {
		file = "unknown"
		line = 0
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// formatLog formats the log message with timestamp, level, caller info, and message
func (l *Logger) formatLog(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelName := levelNames[level]
	callerInfo := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] %s - %s", timestamp, levelName, callerInfo, message)
}

// log writes a log message if the level is sufficient
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level >= l.level {
		l.Logger.Println(l.formatLog(level, format, args...))
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// Package-level logging functions
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, args...)
	}
}
