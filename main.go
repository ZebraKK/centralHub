package main

import (
	"flag"
	"os"

	"github.com/gin-gonic/gin"

	"centralHub/config"
	"centralHub/hubserver"
	"centralHub/logger"
	"centralHub/middleware"
)

func main() {
	// Load configuration
	cfg := loadConfig()

	hubServer := hubserver.NewHubServer()

	router := setupRouter(hubServer, cfg)

	logger.RunLogger.Info().Str("addr", cfg.GetServerAddress()).Msg("centralhub server starting")
	if err := router.Run(cfg.GetServerAddress()); err != nil {
		logger.RunLogger.Fatal().Err(err).Msg("Failed to start centralhub server")
	}
}

// loadConfig loads configuration from file or uses defaults
func loadConfig() *config.Config {
	// Define command line flag for config file path
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Try to load config file
	cfg, err := config.Load(*configPath)
	if err != nil {
		// If config file not found, log warning and use defaults
		logger.RunLogger.Warn().
			Err(err).
			Str("config_path", *configPath).
			Msg("Failed to load config file, using defaults")

		// Return default configuration
		cfg = getDefaultConfig()
	} else {
		logger.RunLogger.Info().
			Str("config_path", *configPath).
			Msg("Configuration loaded successfully")
	}

	return cfg
}

// getDefaultConfig returns a default configuration
func getDefaultConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:    "8080",
			Mode:    "debug",
			Timeout: 30,
		},
		Database: config.DatabaseConfig{
			MongoDB: config.MongoDBConfig{
				URI:      "mongodb://localhost:27017",
				Database: "centralhub",
				Timeout:  10,
			},
		},
		Logger: config.LoggerConfig{
			Level:      "info",
			Output:     "stdout",
			FilePath:   "logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
		},
		External: config.ExternalConfig{
			Volcengine: config.VolcengineConfig{
				AccessKey: "",
				SecretKey: "",
				Region:    "cn-beijing",
			},
		},
	}
}

func setupRouter(hubServer *hubserver.HubServer, cfg *config.Config) *gin.Engine {
	// Initialize logger based on config
	isProd := cfg.IsProduction()
	logger.InitLogger(isProd)

	// Set Gin mode based on config
	ginMode := cfg.Server.Mode
	if ginMode == "" {
		ginMode = os.Getenv("GIN_MODE")
	}
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// Create router with default middleware
	r := gin.Default()

	// authorization
	// metrics
	// tracing
	// recovery
	// Add custom middleware
	r.Use(middleware.AuditLog())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "centralhub",
		})
	})

	// Define routes
	r.POST("/create", hubServer.HandleCreate)
	r.GET("/query", hubServer.HandleQuery)

	return r
}
