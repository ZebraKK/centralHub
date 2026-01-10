package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logger   LoggerConfig   `yaml:"logger"`
	External ExternalConfig `yaml:"external"`
}

// ServerConfig represents server-related configuration
type ServerConfig struct {
	Port    string `yaml:"port"`
	Mode    string `yaml:"mode"` // debug, release, test
	Timeout int    `yaml:"timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	MongoDB MongoDBConfig `yaml:"mongodb"`
}

// MongoDBConfig represents MongoDB connection configuration
type MongoDBConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
	Timeout  int    `yaml:"timeout"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Level      string `yaml:"level"`  // debug, info, warn, error
	Output     string `yaml:"output"` // stdout, file
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`    // megabytes
	MaxBackups int    `yaml:"max_backups"` // number of backups
	MaxAge     int    `yaml:"max_age"`     // days
}

// ExternalConfig represents external service configurations
type ExternalConfig struct {
	Volcengine VolcengineConfig `yaml:"volcengine"`
}

// VolcengineConfig represents Volcengine SDK configuration
type VolcengineConfig struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
}

var GlobalConfig *Config

// Load loads configuration from the specified file path
func Load(configPath string) (*Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
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
