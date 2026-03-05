package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"centralHub/logger"
)

// 配置相关常量
const (
	DefaultConfigPath = "config.json"
	EnvPrefix         = "CH_"
	ReloadInterval    = 30 * time.Second
)

// Config represents the application configuration
type Config struct {
	Server           ServerConfig           `json:"server"`
	Database         DatabaseConfig         `json:"database"`
	Logger           LoggerConfig           `json:"logger"`
	External         ExternalConfig         `json:"external"`
	Features         FeaturesConfig         `json:"features"`
	MockServices     MockServicesConfig     `json:"mock_services"`
	ExternalServices ExternalServicesConfig `json:"external_services"`
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
	Mock       MockConfig       `json:"mock"`
}

// MockServicesConfig represents configurations for all mock services
type MockServicesConfig struct {
	DNS      MockServerCfg `json:"dns"`       // DNS 模拟服务配置
	ICP      MockServerCfg `json:"icp"`       // ICP 模拟服务配置
	CDN      MockServerCfg `json:"cdn"`       // CDN 模拟服务配置
	AdminAPI MockServerCfg `json:"admin_api"` // 管理API服务配置
}

// ExternalServicesConfig represents configurations for all external services
type ExternalServicesConfig struct {
	DNS DNSServiceConfig `json:"dns"` // DNS 服务配置
	ICP ICPServiceConfig `json:"icp"` // ICP 服务配置
	CDN CDNServiceConfig `json:"cdn"` // CDN 服务配置
}

// DNSServiceConfig represents DNS service configuration
type DNSServiceConfig struct {
	APIURL   string `json:"api_url"`  // API 地址
	APIKey   string `json:"api_key"`  // API 密钥
	Provider string `json:"provider"` // 提供商
}

// ICPServiceConfig represents ICP service configuration
type ICPServiceConfig struct {
	APIURL    string `json:"api_url"`    // API 地址
	APIKey    string `json:"api_key"`    // API 密钥
	APISecret string `json:"api_secret"` // API 密钥
}

// CDNServiceConfig represents CDN service configuration
type CDNServiceConfig struct {
	APIURL string `json:"api_url"` // API 地址
	APIKey string `json:"api_key"` // API 密钥
}

// MockConfig 模拟服务配置
type MockConfig struct {
	Enabled  bool                     `json:"enabled"`   // 是否启用模拟服务
	Servers  map[string]MockServerCfg `json:"servers"`   // 模拟服务配置
	BasePath string                   `json:"base_path"` // 模拟服务基础路径
}

// MockServerCfg 模拟服务配置
type MockServerCfg struct {
	Enabled     bool    `json:"enabled"`      // 是否启用
	Port        int     `json:"port"`         // 端口
	Delay       int     `json:"delay"`        // 延迟（毫秒）
	ErrorRate   float64 `json:"error_rate"`   // 错误率（0-1）
	RecordMode  bool    `json:"record_mode"`  // 是否记录请求
	ReplayMode  bool    `json:"replay_mode"`  // 是否回放请求
	RecordsPath string  `json:"records_path"` // 记录文件路径
}

// VolcengineConfig represents Volcengine SDK configuration
type VolcengineConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

// FeaturesConfig represents feature flags and toggles
type FeaturesConfig struct {
	EnableHotReload bool `json:"enable_hot_reload"` // 是否启用配置热重载
	UseMockServices bool `json:"use_mock_services"` // 是否使用模拟服务
	RequireICP      bool `json:"require_icp"`       // 是否要求ICP备案（false时只警告不阻止）
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *Config
	configPath string
	lastMod    time.Time
	mutex      sync.RWMutex
	callbacks  []func(*Config)
}

var (
	// 全局配置管理器
	manager *ConfigManager
	// 全局配置（向后兼容）
	GlobalConfig *Config
)

// 初始化配置管理器
func init() {
	manager = &ConfigManager{
		config:    &Config{},
		callbacks: make([]func(*Config), 0),
	}
}

// Load loads configuration from the specified file path (JSON format)
func Load(configPath string) (*Config, error) {
	// 更新配置管理器
	manager.configPath = configPath

	// 加载配置
	cfg, err := manager.loadConfig()
	if err != nil {
		return nil, err
	}

	// 应用环境变量覆盖
	ApplyEnvironmentOverrides(cfg)

	// 验证配置
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// 设置全局配置（向后兼容）
	GlobalConfig = cfg
	manager.config = cfg

	// 如果启用了热重载，启动后台监控
	if cfg.Features.EnableHotReload {
		go manager.startConfigWatcher()
	}

	return cfg, nil
}

// 加载配置文件
func (cm *ConfigManager) loadConfig() (*Config, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 获取文件修改时间
	fileInfo, err := os.Stat(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get config file info: %w", err)
	}
	cm.lastMod = fileInfo.ModTime()

	// 解析 JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// ApplyEnvironmentOverrides applies environment variable overrides to the configuration
func ApplyEnvironmentOverrides(cfg *Config) {
	// 服务器配置
	if port := os.Getenv(EnvPrefix + "APP_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if mode := os.Getenv(EnvPrefix + "GIN_MODE"); mode != "" {
		cfg.Server.Mode = mode
	} else if mode := os.Getenv("GIN_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}
	if timeout := os.Getenv(EnvPrefix + "SERVER_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			cfg.Server.Timeout = t
		}
	}

	// 数据库配置
	if uri := buildMongoURIFromEnv(); uri != "" {
		cfg.Database.MongoDB.URI = uri
	}
	if database := os.Getenv(EnvPrefix + "MONGO_DATABASE"); database != "" {
		cfg.Database.MongoDB.Database = database
	}
	if timeout := os.Getenv(EnvPrefix + "MONGO_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			cfg.Database.MongoDB.Timeout = t
		}
	}

	// 日志配置
	if level := os.Getenv(EnvPrefix + "LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}
	if output := os.Getenv(EnvPrefix + "LOG_OUTPUT"); output != "" {
		cfg.Logger.Output = output
	}
	if filePath := os.Getenv(EnvPrefix + "LOG_FILE_PATH"); filePath != "" {
		cfg.Logger.FilePath = filePath
	}

	// Volcengine 配置
	if accessKey := os.Getenv(EnvPrefix + "VOLC_ACCESS_KEY"); accessKey != "" {
		cfg.External.Volcengine.AccessKey = accessKey
	}
	if secretKey := os.Getenv(EnvPrefix + "VOLC_SECRET_KEY"); secretKey != "" {
		cfg.External.Volcengine.SecretKey = secretKey
	}
	if region := os.Getenv(EnvPrefix + "VOLC_REGION"); region != "" {
		cfg.External.Volcengine.Region = region
	}

	// 功能特性配置
	if enableHotReload := os.Getenv(EnvPrefix + "ENABLE_HOT_RELOAD"); enableHotReload != "" {
		cfg.Features.EnableHotReload = enableHotReload == "true" || enableHotReload == "1"
	}
	if useMockServices := os.Getenv(EnvPrefix + "USE_MOCK_SERVICES"); useMockServices != "" {
		cfg.Features.UseMockServices = useMockServices == "true" || useMockServices == "1"
	}

	// Mock 服务配置
	if mockEnabled := os.Getenv(EnvPrefix + "MOCK_ENABLED"); mockEnabled != "" {
		cfg.External.Mock.Enabled = mockEnabled == "true" || mockEnabled == "1"
	}
	if mockBasePath := os.Getenv(EnvPrefix + "MOCK_BASE_PATH"); mockBasePath != "" {
		cfg.External.Mock.BasePath = mockBasePath
	}
}

// 从环境变量构建 MongoDB URI
func buildMongoURIFromEnv() string {
	host := os.Getenv(EnvPrefix + "MONGO_HOST")
	port := os.Getenv(EnvPrefix + "MONGO_PORT")
	username := os.Getenv(EnvPrefix + "MONGO_USERNAME")
	password := os.Getenv(EnvPrefix + "MONGO_PASSWORD")

	// 如果没有设置主机，则返回空
	if host == "" {
		return ""
	}

	// 构建 URI
	var uri strings.Builder

	uri.WriteString("mongodb://")

	// 添加认证信息
	if username != "" {
		uri.WriteString(username)
		if password != "" {
			uri.WriteString(":")
			uri.WriteString(password)
		}
		uri.WriteString("@")
	}

	// 添加主机和端口
	uri.WriteString(host)
	if port != "" {
		uri.WriteString(":")
		uri.WriteString(port)
	}

	// 添加数据库名
	database := os.Getenv(EnvPrefix + "MONGO_DATABASE")
	if database != "" {
		uri.WriteString("/")
		uri.WriteString(database)
	}

	return uri.String()
}

// 验证配置
func validateConfig(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug" // 默认模式
	} else if cfg.Server.Mode != "debug" && cfg.Server.Mode != "release" && cfg.Server.Mode != "test" {
		return fmt.Errorf("invalid server mode: %s (must be debug, release, or test)", cfg.Server.Mode)
	}
	if cfg.Server.Timeout <= 0 {
		cfg.Server.Timeout = 30 // 默认超时时间
	}

	// 验证数据库配置
	if cfg.Database.MongoDB.URI == "" {
		return fmt.Errorf("mongodb URI is required")
	}
	if cfg.Database.MongoDB.Database == "" {
		return fmt.Errorf("mongodb database name is required")
	}
	if cfg.Database.MongoDB.Timeout <= 0 {
		cfg.Database.MongoDB.Timeout = 10 // 默认超时时间
	}

	// 验证日志配置
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info" // 默认级别
	} else if !isValidLogLevel(cfg.Logger.Level) {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", cfg.Logger.Level)
	}
	if cfg.Logger.Output == "" {
		cfg.Logger.Output = "stdout" // 默认输出
	} else if cfg.Logger.Output != "stdout" && cfg.Logger.Output != "file" {
		return fmt.Errorf("invalid log output: %s (must be stdout or file)", cfg.Logger.Output)
	}
	if cfg.Logger.Output == "file" && cfg.Logger.FilePath == "" {
		return fmt.Errorf("log file path is required when output is file")
	}
	if cfg.Logger.MaxSize <= 0 {
		cfg.Logger.MaxSize = 100 // 默认大小
	}
	if cfg.Logger.MaxBackups <= 0 {
		cfg.Logger.MaxBackups = 3 // 默认备份数
	}
	if cfg.Logger.MaxAge <= 0 {
		cfg.Logger.MaxAge = 7 // 默认保留天数
	}

	// 验证 Mock 配置
	if cfg.External.Mock.Enabled {
		if cfg.External.Mock.BasePath == "" {
			cfg.External.Mock.BasePath = "mock/data" // 默认路径
		}
		if cfg.External.Mock.Servers == nil {
			cfg.External.Mock.Servers = make(map[string]MockServerCfg)
		}

		// 验证每个 Mock 服务配置
		for name, serverCfg := range cfg.External.Mock.Servers {
			if serverCfg.Enabled {
				if serverCfg.Port <= 0 {
					return fmt.Errorf("invalid port for mock server %s: %d", name, serverCfg.Port)
				}
				if serverCfg.Delay < 0 {
					serverCfg.Delay = 0
				}
				if serverCfg.ErrorRate < 0 || serverCfg.ErrorRate > 1 {
					return fmt.Errorf("invalid error rate for mock server %s: %f (must be between 0 and 1)", name, serverCfg.ErrorRate)
				}
				if serverCfg.RecordsPath == "" {
					serverCfg.RecordsPath = fmt.Sprintf("%s/%s", cfg.External.Mock.BasePath, name)
				}
				cfg.External.Mock.Servers[name] = serverCfg
			}
		}
	}

	// 验证 MockServices 配置
	if cfg.Features.UseMockServices {
		// 设置 DNS 模拟服务默认配置
		if cfg.MockServices.DNS.Port <= 0 {
			cfg.MockServices.DNS.Port = 8091 // 默认端口
		}
		if cfg.MockServices.DNS.Delay < 0 {
			cfg.MockServices.DNS.Delay = 0
		}
		if cfg.MockServices.DNS.ErrorRate < 0 || cfg.MockServices.DNS.ErrorRate > 1 {
			return fmt.Errorf("invalid error rate for DNS mock server: %f (must be between 0 and 1)", cfg.MockServices.DNS.ErrorRate)
		}
		if cfg.MockServices.DNS.RecordsPath == "" {
			cfg.MockServices.DNS.RecordsPath = "mock/data/dns"
		}

		// 设置 ICP 模拟服务默认配置
		if cfg.MockServices.ICP.Port <= 0 {
			cfg.MockServices.ICP.Port = 8092 // 默认端口
		}

		// 设置 CDN 模拟服务默认配置
		if cfg.MockServices.CDN.Port <= 0 {
			cfg.MockServices.CDN.Port = 8093 // 默认端口
		}

		// 设置管理API服务默认配置
		if cfg.MockServices.AdminAPI.Port <= 0 {
			cfg.MockServices.AdminAPI.Port = 8090 // 默认端口
		}
	}

	// 验证 ExternalServices 配置
	if !cfg.Features.UseMockServices {
		if cfg.ExternalServices.DNS.APIURL == "" {
			cfg.ExternalServices.DNS.APIURL = "https://api.example.com/dns" // 默认API地址
		}
		if cfg.ExternalServices.DNS.Provider == "" {
			cfg.ExternalServices.DNS.Provider = "default" // 默认提供商
		}
	}

	return nil
}

// 检查日志级别是否有效
func isValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error"}
	for _, l := range validLevels {
		if level == l {
			return true
		}
	}
	return false
}

// 启动配置文件监控
func (cm *ConfigManager) startConfigWatcher() {
	ticker := time.NewTicker(ReloadInterval)
	defer ticker.Stop()

	for range ticker.C {
		if cm.checkAndReloadConfig() {
			logger.RunLogger.Info().Str("path", cm.configPath).Msg("Configuration reloaded")
		}
	}
}

// 检查并重新加载配置
func (cm *ConfigManager) checkAndReloadConfig() bool {
	// 获取文件信息
	fileInfo, err := os.Stat(cm.configPath)
	if err != nil {
		logger.RunLogger.Error().Err(err).Str("path", cm.configPath).Msg("Failed to get config file info")
		return false
	}

	// 检查文件是否被修改
	if fileInfo.ModTime().After(cm.lastMod) {
		// 加载新配置
		newCfg, err := cm.loadConfig()
		if err != nil {
			logger.RunLogger.Error().Err(err).Str("path", cm.configPath).Msg("Failed to reload config")
			return false
		}

		// 应用环境变量覆盖
		ApplyEnvironmentOverrides(newCfg)

		// 验证配置
		if err := validateConfig(newCfg); err != nil {
			logger.RunLogger.Error().Err(err).Str("path", cm.configPath).Msg("Invalid configuration")
			return false
		}

		// 更新全局配置
		cm.mutex.Lock()
		GlobalConfig = newCfg
		cm.config = newCfg
		cm.mutex.Unlock()

		// 调用回调函数
		for _, callback := range cm.callbacks {
			go callback(newCfg)
		}

		return true
	}

	return false
}

// RegisterCallback 注册配置变更回调函数
func RegisterCallback(callback func(*Config)) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.callbacks = append(manager.callbacks, callback)
}

// GetConfig 获取当前配置（线程安全）
func GetConfig() *Config {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.config
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return ":" + c.Server.Port
}

// IsProduction returns whether the server is running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}

// ShouldUseMockServices returns whether mock services should be used
func (c *Config) ShouldUseMockServices() bool {
	return c.Features.UseMockServices
}
