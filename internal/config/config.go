package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	ServerPort string
	ServerHost string
	ServerMode string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT configuration
	JWTSecret     string
	JWTExpiration string

	// Logging configuration
	LogLevel string
	LogFile  string

	// API configuration
	APIVersion string
	APIPrefix  string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		// Server configuration
		ServerPort: getEnvOrDefault("SERVER_PORT", "8080"),
		ServerHost: getEnvOrDefault("SERVER_HOST", "localhost"),
		ServerMode: getEnvOrDefault("SERVER_MODE", "debug"),

		// Database configuration
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "3306"),
		DBUser:     getEnvOrDefault("DB_USER", "root"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:     getEnvOrDefault("DB_NAME", "app_db"),

		// JWT configuration
		JWTSecret:     getEnvOrDefault("JWT_SECRET", "your-secret-key"),
		JWTExpiration: getEnvOrDefault("JWT_EXPIRATION", "24h"),

		// Logging configuration
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
		LogFile:  getEnvOrDefault("LOG_FILE", "app.log"),

		// API configuration
		APIVersion: getEnvOrDefault("API_VERSION", "v1"),
		APIPrefix:  getEnvOrDefault("API_PREFIX", "/api"),
	}

	// Validate required configuration
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks if all required configuration is present
func (c *Config) validate() error {
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	return nil
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.ServerHost, c.ServerPort)
}

// GetAPIPath returns the full API path
func (c *Config) GetAPIPath() string {
	return fmt.Sprintf("%s/%s", c.APIPrefix, c.APIVersion)
}
