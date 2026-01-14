package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Logger   LoggerConfig   `json:"logger"`
	External ExternalConfig `json:"external"`
}

// ServerConfig represents server-related configuration
type ServerConfig struct {
	Port    string `json:"port"`
	Mode    string `json:"mode"` // debug, release, test
	Timeout int    `json:"timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	MongoDB MongoDBConfig `json:"mongodb"`
}

// MongoDBConfig represents MongoDB connection configuration
type MongoDBConfig struct {
	URI      string `json:"uri"`
	Database string `json:"database"`
	Timeout  int    `json:"timeout"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Level      string `json:"level"`  // debug, info, warn, error
	Output     string `json:"output"` // stdout, file
	FilePath   string `json:"file_path"`
	MaxSize    int    `json:"max_size"`    // megabytes
	MaxBackups int    `json:"max_backups"` // number of backups
	MaxAge     int    `json:"max_age"`     // days
}

// ExternalConfig represents external service configurations
type ExternalConfig struct {
	Volcengine VolcengineConfig `json:"volcengine"`
}

// VolcengineConfig represents Volcengine SDK configuration
type VolcengineConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

var GlobalConfig *Config

// Load loads configuration from the specified file path (JSON format)
func Load(configPath string) (*Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set global config
	GlobalConfig = &cfg

	return &cfg, nil
}

// validate validates the configuration
func (c *Config) validate() error {
	// Validate server config
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug" // default mode
	}

	// Validate database config
	if c.Database.MongoDB.URI == "" {
		return fmt.Errorf("mongodb URI is required")
	}
	if c.Database.MongoDB.Database == "" {
		return fmt.Errorf("mongodb database name is required")
	}

	// Validate logger config
	if c.Logger.Level == "" {
		c.Logger.Level = "info" // default level
	}
	if c.Logger.Output == "" {
		c.Logger.Output = "stdout" // default output
	}

	return nil
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return ":" + c.Server.Port
}

// IsProduction returns whether the server is running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}
