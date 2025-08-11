package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the auction microservice
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	Metrics  MetricsConfig  `json:"metrics"`
	Logging  LoggingConfig  `json:"logging"`
	WebSocket WebSocketConfig `json:"websocket"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Database     string `json:"database"`
	SSLMode      string `json:"ssl_mode"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SecretKey      string        `json:"secret_key"`
	ExpirationTime time.Duration `json:"expiration_time"`
	Issuer         string        `json:"issuer"`
}

// MetricsConfig holds metrics-related configuration
type MetricsConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
	Path    string `json:"path"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"` // "json" or "console"
}

// WebSocketConfig holds WebSocket-related configuration
type WebSocketConfig struct {
	ReadBufferSize  int           `json:"read_buffer_size"`
	WriteBufferSize int           `json:"write_buffer_size"`
	WriteWait       time.Duration `json:"write_wait"`
	PongWait        time.Duration `json:"pong_wait"`
	PingPeriod      time.Duration `json:"ping_period"`
	MaxMessageSize  int64         `json:"max_message_size"`
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "auction_user"),
			Password:     getEnv("DB_PASSWORD", "auction_password"),
			Database:     getEnv("DB_NAME", "auction_db"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		},
		JWT: JWTConfig{
			SecretKey:      getEnv("JWT_SECRET_KEY", "your-secret-key-change-this-in-production"),
			ExpirationTime: getEnvAsDuration("JWT_EXPIRATION_TIME", 24*time.Hour),
			Issuer:         getEnv("JWT_ISSUER", "auction-microservice"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED", true),
			Port:    getEnvAsInt("METRICS_PORT", 9090),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize: getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
			WriteWait:       getEnvAsDuration("WS_WRITE_WAIT", 10*time.Second),
			PongWait:        getEnvAsDuration("WS_PONG_WAIT", 60*time.Second),
			PingPeriod:      getEnvAsDuration("WS_PING_PERIOD", 54*time.Second),
			MaxMessageSize:  getEnvAsInt64("WS_MAX_MESSAGE_SIZE", 512),
		},
	}
}

// LoadForTesting loads configuration optimized for testing
func LoadForTesting() *Config {
	config := Load()
	
	// Override with testing-specific values
	config.Database.Host = "localhost"
	config.Database.Database = "auction_test_db"
	config.JWT.SecretKey = "test-secret-key"
	config.Logging.Level = "debug"
	config.Logging.Format = "console"
	config.Metrics.Enabled = false
	
	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Add validation logic here if needed
	if c.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}
	
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	
	return nil
}

// Helper functions to read environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}