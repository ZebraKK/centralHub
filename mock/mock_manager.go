package mock

import (
	"fmt"
	"sync"

	"centralHub/config"
	"centralHub/logger"
)

// MockServerManager 模拟服务管理器
type MockServerManager struct {
	servers map[string]MockServer
	mutex   sync.RWMutex
	config  *config.Config
}

// NewMockServerManager 创建模拟服务管理器
func NewMockServerManager(cfg *config.Config) *MockServerManager {
	return &MockServerManager{
		servers: make(map[string]MockServer),
		config:  cfg,
	}
}

// RegisterServer 注册模拟服务
func (m *MockServerManager) RegisterServer(name string, server MockServer) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.servers[name]; exists {
		return fmt.Errorf("server already registered: %s", name)
	}

	m.servers[name] = server
	logger.RunLogger.Info().
		Str("server", name).
		Msg("Mock server registered")

	return nil
}

// GetServer 获取模拟服务
func (m *MockServerManager) GetServer(name string) (MockServer, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	server, exists := m.servers[name]
	if !exists {
		return nil, fmt.Errorf("server not found: %s", name)
	}

	return server, nil
}

// StartServer 启动模拟服务
func (m *MockServerManager) StartServer(name string) error {
	server, err := m.GetServer(name)
	if err != nil {
		return err
	}

	if err := server.Start(); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("server", name).
			Msg("Failed to start mock server")
		return err
	}

	logger.RunLogger.Info().
		Str("server", name).
		Msg("Mock server started")

	return nil
}

// StopServer 停止模拟服务
func (m *MockServerManager) StopServer(name string) error {
	server, err := m.GetServer(name)
	if err != nil {
		return err
	}

	if err := server.Stop(); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("server", name).
			Msg("Failed to stop mock server")
		return err
	}

	logger.RunLogger.Info().
		Str("server", name).
		Msg("Mock server stopped")

	return nil
}

// StartAllServers 启动所有模拟服务
func (m *MockServerManager) StartAllServers() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, server := range m.servers {
		if err := server.Start(); err != nil {
			logger.RunLogger.Error().
				Err(err).
				Str("server", name).
				Msg("Failed to start mock server")
			continue
		}

		logger.RunLogger.Info().
			Str("server", name).
			Msg("Mock server started")
	}

	return nil
}

// StopAllServers 停止所有模拟服务
func (m *MockServerManager) StopAllServers() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, server := range m.servers {
		if err := server.Stop(); err != nil {
			logger.RunLogger.Error().
				Err(err).
				Str("server", name).
				Msg("Failed to stop mock server")
			continue
		}

		logger.RunLogger.Info().
			Str("server", name).
			Msg("Mock server stopped")
	}

	return nil
}

// GetServerNames 获取所有模拟服务名称
func (m *MockServerManager) GetServerNames() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.servers))
	for name := range m.servers {
		names = append(names, name)
	}

	return names
}

// GetServerConfigs 获取所有模拟服务配置
func (m *MockServerManager) GetServerConfigs() map[string]MockConfig {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	configs := make(map[string]MockConfig)
	for name, server := range m.servers {
		configs[name] = server.GetConfig()
	}

	return configs
}

// SetServerConfig 设置模拟服务配置
func (m *MockServerManager) SetServerConfig(name string, config MockConfig) error {
	server, err := m.GetServer(name)
	if err != nil {
		return err
	}

	if err := server.SetConfig(config); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("server", name).
			Msg("Failed to set mock server config")
		return err
	}

	logger.RunLogger.Info().
		Str("server", name).
		Msg("Mock server config updated")

	return nil
}

// SaveServerRecords 保存模拟服务请求记录
func (m *MockServerManager) SaveServerRecords(name, path string) error {
	server, err := m.GetServer(name)
	if err != nil {
		return err
	}

	if err := server.SaveRecords(path); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("server", name).
			Str("path", path).
			Msg("Failed to save mock server records")
		return err
	}

	logger.RunLogger.Info().
		Str("server", name).
		Str("path", path).
		Msg("Mock server records saved")

	return nil
}

// LoadServerRecords 加载模拟服务请求记录
func (m *MockServerManager) LoadServerRecords(name, path string) error {
	server, err := m.GetServer(name)
	if err != nil {
		return err
	}

	if err := server.LoadRecords(path); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("server", name).
			Str("path", path).
			Msg("Failed to load mock server records")
		return err
	}

	logger.RunLogger.Info().
		Str("server", name).
		Str("path", path).
		Msg("Mock server records loaded")

	return nil
}

// ClearServerRecords 清除模拟服务请求记录
func (m *MockServerManager) ClearServerRecords(name string) error {
	server, err := m.GetServer(name)
	if err != nil {
		return err
	}

	server.ClearRecords()

	logger.RunLogger.Info().
		Str("server", name).
		Msg("Mock server records cleared")

	return nil
}

// GetServerRecords 获取模拟服务请求记录
func (m *MockServerManager) GetServerRecords(name string) ([]RequestRecord, error) {
	server, err := m.GetServer(name)
	if err != nil {
		return nil, err
	}

	return server.GetRecords(), nil
}
