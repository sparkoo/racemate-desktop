package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds configuration for the logger
type Config struct {
	LogsDir    string
	LogFile    string
	MaxSize    int  // megabytes
	MaxBackups int  // number of backups
	MaxAge     int  // days
	Compress   bool // compress rotated logs
}

// DefaultConfig returns a default logger configuration
func DefaultConfig(logsDir string) Config {
	return Config{
		LogsDir:    logsDir,
		LogFile:    "racemate.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
}

// Initialize creates and configures a new structured logger
func Initialize(config Config) *slog.Logger {
	// Configure log file with rotation
	logFilePath := filepath.Join(config.LogsDir, config.LogFile)
	fileLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Create a multi-writer that writes to both stdout and the log file
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	// Configure slog options
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// Use JSON format for structured logging
	handler := slog.NewJSONHandler(multiWriter, opts)

	// Create the logger
	logger := slog.New(handler)

	// Log startup message
	logger.Info("Logger initialized",
		"logFile", logFilePath,
		"maxSize", config.MaxSize,
		"maxBackups", config.MaxBackups,
		"maxAge", config.MaxAge,
		"compress", config.Compress)

	// Create a custom logger adapter for the standard log package
	// This ensures that log.Printf, log.Println, etc. will use our structured logger
	log.SetFlags(0) // Remove default timestamps as slog will add them
	log.SetOutput(&logAdapter{logger: logger})

	return logger
}

// logAdapter is a custom io.Writer that forwards standard log package writes to slog
type logAdapter struct {
	logger *slog.Logger
}

// Write implements io.Writer for the logAdapter
func (a *logAdapter) Write(p []byte) (n int, err error) {
	// Remove trailing newlines for cleaner log output
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	a.logger.Info(msg)

	// Return the original length to satisfy io.Writer
	return len(p), nil
}
