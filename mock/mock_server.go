package mock

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"centralHub/logger"
)

// MockConfig 模拟服务配置
type MockConfig struct {
	Enabled     bool    `json:"enabled"`      // 是否启用模拟服务
	Port        int     `json:"port"`         // 模拟服务端口
	Delay       int     `json:"delay"`        // 模拟延迟（毫秒）
	ErrorRate   float64 `json:"error_rate"`   // 错误率（0-1）
	RecordMode  bool    `json:"record_mode"`  // 是否记录请求
	ReplayMode  bool    `json:"replay_mode"`  // 是否回放请求
	RecordsPath string  `json:"records_path"` // 记录文件路径
}

// DefaultMockConfig 默认模拟服务配置
func DefaultMockConfig() MockConfig {
	return MockConfig{
		Enabled:     true,
		Port:        8090,
		Delay:       100,
		ErrorRate:   0.0,
		RecordMode:  false,
		ReplayMode:  false,
		RecordsPath: "mock/records",
	}
}

// MockResponse 模拟响应
type MockResponse struct {
	StatusCode int               `json:"status_code"` // HTTP状态码
	Headers    map[string]string `json:"headers"`     // 响应头
	Body       interface{}       `json:"body"`        // 响应体
	Delay      int               `json:"delay"`       // 延迟（毫秒）
}

// RequestRecord 请求记录
type RequestRecord struct {
	Timestamp  int64             `json:"timestamp"`   // 时间戳
	Method     string            `json:"method"`      // HTTP方法
	Path       string            `json:"path"`        // 请求路径
	Query      map[string]string `json:"query"`       // 查询参数
	Headers    map[string]string `json:"headers"`     // 请求头
	Body       interface{}       `json:"body"`        // 请求体
	Response   MockResponse      `json:"response"`    // 响应
	StatusCode int               `json:"status_code"` // 状态码
}

// MockServer 模拟服务接口
type MockServer interface {
	Start() error                                          // 启动服务
	Stop() error                                           // 停止服务
	SetConfig(config MockConfig) error                     // 设置配置
	GetConfig() MockConfig                                 // 获取配置
	AddRoute(method, path string, handler gin.HandlerFunc) // 添加路由
	GetRecords() []RequestRecord                           // 获取请求记录
	ClearRecords()                                         // 清除请求记录
	SaveRecords(path string) error                         // 保存请求记录
	LoadRecords(path string) error                         // 加载请求记录
}

// BaseMockServer 基础模拟服务实现
type BaseMockServer struct {
	Config     MockConfig                // 配置
	Router     *gin.Engine               // 路由
	Server     *http.Server              // HTTP服务器
	Records    []RequestRecord           // 请求记录
	RecordsMux sync.RWMutex              // 请求记录互斥锁
	Templates  map[string][]MockResponse // 响应模板
	Running    bool                      // 是否运行中
	ServerType string                    // 服务类型
}

// NewBaseMockServer 创建基础模拟服务
func NewBaseMockServer(serverType string, config MockConfig) *BaseMockServer {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 创建Gin路由
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 创建模拟服务
	server := &BaseMockServer{
		Config:     config,
		Router:     router,
		Records:    make([]RequestRecord, 0),
		Templates:  make(map[string][]MockResponse),
		ServerType: serverType,
	}

	// 添加管理API
	server.setupAdminAPI()

	return server
}

// setupAdminAPI 设置管理API
func (s *BaseMockServer) setupAdminAPI() {
	admin := s.Router.Group("/admin")
	{
		// 获取配置
		admin.GET("/config", func(c *gin.Context) {
			c.JSON(http.StatusOK, s.Config)
		})

		// 更新配置
		admin.POST("/config", func(c *gin.Context) {
			var config MockConfig
			if err := c.ShouldBindJSON(&config); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			s.Config = config
			c.JSON(http.StatusOK, gin.H{"message": "Config updated"})
		})

		// 获取请求记录
		admin.GET("/records", func(c *gin.Context) {
			s.RecordsMux.RLock()
			defer s.RecordsMux.RUnlock()
			c.JSON(http.StatusOK, s.Records)
		})

		// 清除请求记录
		admin.DELETE("/records", func(c *gin.Context) {
			s.ClearRecords()
			c.JSON(http.StatusOK, gin.H{"message": "Records cleared"})
		})

		// 保存请求记录
		admin.POST("/records/save", func(c *gin.Context) {
			var req struct {
				Path string `json:"path"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if req.Path == "" {
				req.Path = s.Config.RecordsPath
			}
			if err := s.SaveRecords(req.Path); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Records saved"})
		})

		// 加载请求记录
		admin.POST("/records/load", func(c *gin.Context) {
			var req struct {
				Path string `json:"path"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if req.Path == "" {
				req.Path = s.Config.RecordsPath
			}
			if err := s.LoadRecords(req.Path); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Records loaded"})
		})

		// 获取响应模板
		admin.GET("/templates", func(c *gin.Context) {
			c.JSON(http.StatusOK, s.Templates)
		})

		// 添加响应模板
		admin.POST("/templates/:key", func(c *gin.Context) {
			key := c.Param("key")
			var templates []MockResponse
			if err := c.ShouldBindJSON(&templates); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			s.Templates[key] = templates
			c.JSON(http.StatusOK, gin.H{"message": "Templates added"})
		})

		// 删除响应模板
		admin.DELETE("/templates/:key", func(c *gin.Context) {
			key := c.Param("key")
			delete(s.Templates, key)
			c.JSON(http.StatusOK, gin.H{"message": "Templates deleted"})
		})

		// 获取服务状态
		admin.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"type":    s.ServerType,
				"running": s.Running,
				"records": len(s.Records),
			})
		})
	}
}

// Start 启动服务
func (s *BaseMockServer) Start() error {
	if s.Running {
		return fmt.Errorf("server already running")
	}

	// 创建HTTP服务器
	s.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Config.Port),
		Handler: s.Router,
	}

	// 启动服务器
	go func() {
		logger.RunLogger.Info().
			Str("type", s.ServerType).
			Int("port", s.Config.Port).
			Msg("Starting mock server")

		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.RunLogger.Error().
				Err(err).
				Str("type", s.ServerType).
				Msg("Failed to start mock server")
		}
	}()

	s.Running = true
	return nil
}

// Stop 停止服务
func (s *BaseMockServer) Stop() error {
	if !s.Running {
		return fmt.Errorf("server not running")
	}

	// 停止服务器
	if s.Server != nil {
		logger.RunLogger.Info().
			Str("type", s.ServerType).
			Msg("Stopping mock server")

		if err := s.Server.Close(); err != nil {
			logger.RunLogger.Error().
				Err(err).
				Str("type", s.ServerType).
				Msg("Failed to stop mock server")
			return err
		}
	}

	s.Running = false
	return nil
}

// SetConfig 设置配置
func (s *BaseMockServer) SetConfig(config MockConfig) error {
	// 如果服务正在运行，需要重启服务
	if s.Running {
		if err := s.Stop(); err != nil {
			return err
		}
		s.Config = config
		return s.Start()
	}

	s.Config = config
	return nil
}

// GetConfig 获取配置
func (s *BaseMockServer) GetConfig() MockConfig {
	return s.Config
}

// AddRoute 添加路由
func (s *BaseMockServer) AddRoute(method, path string, handler gin.HandlerFunc) {
	// 添加中间件，用于记录请求和模拟延迟
	wrappedHandler := s.wrapHandler(handler)

	// 添加路由
	switch method {
	case http.MethodGet:
		s.Router.GET(path, wrappedHandler)
	case http.MethodPost:
		s.Router.POST(path, wrappedHandler)
	case http.MethodPut:
		s.Router.PUT(path, wrappedHandler)
	case http.MethodDelete:
		s.Router.DELETE(path, wrappedHandler)
	case http.MethodPatch:
		s.Router.PATCH(path, wrappedHandler)
	case http.MethodHead:
		s.Router.HEAD(path, wrappedHandler)
	case http.MethodOptions:
		s.Router.OPTIONS(path, wrappedHandler)
	default:
		s.Router.Any(path, wrappedHandler)
	}
}

// wrapHandler 包装处理函数，添加记录和延迟功能
func (s *BaseMockServer) wrapHandler(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 记录请求
		record := RequestRecord{
			Timestamp: startTime.Unix(),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Query:     make(map[string]string),
			Headers:   make(map[string]string),
		}

		// 记录查询参数
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				record.Query[k] = v[0]
			}
		}

		// 记录请求头
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				record.Headers[k] = v[0]
			}
		}

		// 记录请求体
		if c.Request.Body != nil {
			var body interface{}
			if err := c.ShouldBindJSON(&body); err == nil {
				record.Body = body
			}
		}

		// 模拟延迟
		delay := s.Config.Delay
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		// 模拟错误
		if s.Config.ErrorRate > 0 && rand.Float64() < s.Config.ErrorRate {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Mock server error"})
			record.StatusCode = http.StatusInternalServerError
			s.recordRequest(record)
			return
		}

		// 创建自定义响应写入器，用于记录响应
		writer := &responseWriter{ResponseWriter: c.Writer}
		c.Writer = writer

		// 执行原始处理函数
		handler(c)

		// 记录响应
		record.StatusCode = writer.statusCode
		record.Response = MockResponse{
			StatusCode: writer.statusCode,
			Headers:    make(map[string]string),
			Body:       writer.body,
			Delay:      int(time.Since(startTime).Milliseconds()),
		}

		// 记录响应头
		for k, v := range writer.Header() {
			if len(v) > 0 {
				record.Response.Headers[k] = v[0]
			}
		}

		// 保存请求记录
		s.recordRequest(record)
	}
}

// recordRequest 记录请求
func (s *BaseMockServer) recordRequest(record RequestRecord) {
	if !s.Config.RecordMode {
		return
	}

	s.RecordsMux.Lock()
	defer s.RecordsMux.Unlock()
	s.Records = append(s.Records, record)
}

// GetRecords 获取请求记录
func (s *BaseMockServer) GetRecords() []RequestRecord {
	s.RecordsMux.RLock()
	defer s.RecordsMux.RUnlock()
	return s.Records
}

// ClearRecords 清除请求记录
func (s *BaseMockServer) ClearRecords() {
	s.RecordsMux.Lock()
	defer s.RecordsMux.Unlock()
	s.Records = make([]RequestRecord, 0)
}

// SaveRecords 保存请求记录
func (s *BaseMockServer) SaveRecords(path string) error {
	s.RecordsMux.RLock()
	defer s.RecordsMux.RUnlock()

	// 序列化记录
	data, err := json.MarshalIndent(s.Records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal records: %w", err)
	}

	// 保存到文件
	if err := saveToFile(path, data); err != nil {
		return fmt.Errorf("save to file: %w", err)
	}

	return nil
}

// LoadRecords 加载请求记录
func (s *BaseMockServer) LoadRecords(path string) error {
	// 从文件加载
	data, err := loadFromFile(path)
	if err != nil {
		return fmt.Errorf("load from file: %w", err)
	}

	// 反序列化记录
	var records []RequestRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("unmarshal records: %w", err)
	}

	// 更新记录
	s.RecordsMux.Lock()
	defer s.RecordsMux.Unlock()
	s.Records = records

	return nil
}

// GetRandomResponse 获取随机响应
func (s *BaseMockServer) GetRandomResponse(key string) (MockResponse, bool) {
	templates, ok := s.Templates[key]
	if !ok || len(templates) == 0 {
		return MockResponse{}, false
	}

	// 随机选择一个响应
	index := rand.Intn(len(templates))
	return templates[index], true
}

// responseWriter 自定义响应写入器，用于记录响应
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
	body       interface{}
}

// WriteHeader 重写WriteHeader方法，记录状态码
func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write 重写Write方法，记录响应体
func (w *responseWriter) Write(data []byte) (int, error) {
	// 尝试解析JSON
	var body interface{}
	if err := json.Unmarshal(data, &body); err == nil {
		w.body = body
	} else {
		w.body = string(data)
	}
	return w.ResponseWriter.Write(data)
}

// saveToFile 保存数据到文件
func saveToFile(path string, data []byte) error {
	// 创建目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// loadFromFile 从文件加载数据
func loadFromFile(path string) ([]byte, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return data, nil
}
