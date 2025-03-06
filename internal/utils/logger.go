package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cam-boltnote/go-ignite/internal/config"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger to provide application-specific logging
type Logger struct {
	logger zerolog.Logger
}

var (
	defaultLogger *Logger
)

// LogLevel represents available logging levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// InitLogger creates and configures the default logger using application config
func InitLogger(cfg *config.Config) error {
	// Get log level from config
	levelStr := cfg.LogLevel
	if levelStr == "" {
		levelStr = string(InfoLevel) // Default to info level
	}

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0744); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Create a logger instance
	logger, err := NewLogger(LogLevel(levelStr))
	if err != nil {
		return err
	}

	defaultLogger = logger
	return nil
}

// NewLogger creates a new logger instance with the specified level
func NewLogger(level LogLevel) (*Logger, error) {
	// Parse the log level
	zerologLevel, err := zerolog.ParseLevel(string(level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %v", err)
	}

	// Create separate log files for each level
	outputs := make([]zerolog.Level, 0)
	levels := []zerolog.Level{
		zerolog.DebugLevel,
		zerolog.InfoLevel,
		zerolog.WarnLevel,
		zerolog.ErrorLevel,
		zerolog.FatalLevel,
	}

	// Create multi-writer for console and file output
	writers := make([]io.Writer, 0)

	// Add console writer
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	writers = append(writers, consoleWriter)

	// Create level-specific log files
	for _, lvl := range levels {
		if lvl >= zerologLevel {
			outputs = append(outputs, lvl)
			filename := filepath.Join("logs", fmt.Sprintf("%s.log", lvl.String()))
			file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file %s: %v", filename, err)
			}
			writers = append(writers, file)
		}
	}

	// Create the logger with all writers
	logger := zerolog.New(zerolog.MultiLevelWriter(writers...)).
		Level(zerologLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{
		logger: logger,
	}, nil
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	return defaultLogger
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error logs an error message with optional fields
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event.Err(err)
	}
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message with optional fields and exits
func (l *Logger) Fatal(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event.Err(err)
	}
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

// WithService adds service name context to the logger
func (l *Logger) WithService(serviceName string) *Logger {
	newLogger := l.logger.With().Str("service", serviceName).Logger()
	return &Logger{logger: newLogger}
}

// Example usage in a service:
/*
type UserService struct {
    logger *utils.Logger
    db     *connectors.Database
}

func NewUserService(db *connectors.Database) *UserService {
    return &UserService{
        logger: utils.GetLogger().WithService("user_service"),
        db:     db,
    }
}

func (s *UserService) CreateUser(user *models.User) error {
    s.logger.Info("Creating new user", map[string]interface{}{
        "email": user.Email,
    })

    if err := s.db.Create(user).Error; err != nil {
        s.logger.Error("Failed to create user", err, map[string]interface{}{
            "email": user.Email,
        })
        return err
    }

    return nil
}
*/
