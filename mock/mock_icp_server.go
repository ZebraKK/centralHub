package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"centralHub/logger"
	"centralHub/model"
)

// MockICPServer ICP 模拟服务
type MockICPServer struct {
	*BaseMockServer
	icpRecords      map[string]model.ICPRecord
	icpRecordsMutex sync.RWMutex
}

// NewMockICPServer 创建 ICP 模拟服务
func NewMockICPServer(config MockConfig) *MockICPServer {
	baseServer := NewBaseMockServer("icp", config)
	server := &MockICPServer{
		BaseMockServer: baseServer,
		icpRecords:     make(map[string]model.ICPRecord),
	}

	// 添加 ICP 查询路由
	server.setupRoutes()

	return server
}

// setupRoutes 设置路由
func (s *MockICPServer) setupRoutes() {
	// ICP 管理 API
	icp := s.Router.Group("/icp")
	{
		// 查询 ICP 备案信息
		icp.GET("/query", s.handleICPQuery)

		// 管理 API
		admin := icp.Group("/admin")
		{
			// 获取所有 ICP 记录
			admin.GET("/records", s.handleGetAllICPRecords)

			// 添加 ICP 记录
			admin.POST("/records", s.handleAddICPRecord)

			// 删除 ICP 记录
			admin.DELETE("/records/:domain", s.handleDeleteICPRecord)

			// 批量导入 ICP 记录
			admin.POST("/records/import", s.handleImportICPRecords)

			// 清空所有 ICP 记录
			admin.DELETE("/records", s.handleClearICPRecords)
		}
	}
}

// handleICPQuery 处理 ICP 查询请求
func (s *MockICPServer) handleICPQuery(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter is required"})
		return
	}

	// 规范化域名
	domain = strings.ToLower(strings.TrimSpace(domain))

	// 查询 ICP 记录
	s.icpRecordsMutex.RLock()
	record, exists := s.icpRecords[domain]
	s.icpRecordsMutex.RUnlock()

	if !exists {
		// 尝试查找主域名
		parts := strings.Split(domain, ".")
		if len(parts) > 2 {
			mainDomain := parts[len(parts)-2] + "." + parts[len(parts)-1]
			s.icpRecordsMutex.RLock()
			record, exists = s.icpRecords[mainDomain]
			s.icpRecordsMutex.RUnlock()
		}
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "ICP record not found"})
		return
	}

	c.JSON(http.StatusOK, record)
}

// handleGetAllICPRecords 处理获取所有 ICP 记录请求
func (s *MockICPServer) handleGetAllICPRecords(c *gin.Context) {
	s.icpRecordsMutex.RLock()
	defer s.icpRecordsMutex.RUnlock()

	records := make([]model.ICPRecord, 0, len(s.icpRecords))
	for _, record := range s.icpRecords {
		records = append(records, record)
	}

	c.JSON(http.StatusOK, records)
}

// handleAddICPRecord 处理添加 ICP 记录请求
func (s *MockICPServer) handleAddICPRecord(c *gin.Context) {
	var record model.ICPRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证记录
	if record.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain is required"})
		return
	}

	// 规范化域名
	record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))

	// 添加记录
	s.icpRecordsMutex.Lock()
	s.icpRecords[record.Domain] = record
	s.icpRecordsMutex.Unlock()

	logger.RunLogger.Info().
		Str("domain", record.Domain).
		Str("icp_number", record.ICPNumber).
		Msg("Added ICP record")

	c.JSON(http.StatusOK, gin.H{"message": "ICP record added"})
}

// handleDeleteICPRecord 处理删除 ICP 记录请求
func (s *MockICPServer) handleDeleteICPRecord(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter is required"})
		return
	}

	// 规范化域名
	domain = strings.ToLower(strings.TrimSpace(domain))

	// 删除记录
	s.icpRecordsMutex.Lock()
	delete(s.icpRecords, domain)
	s.icpRecordsMutex.Unlock()

	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Deleted ICP record")

	c.JSON(http.StatusOK, gin.H{"message": "ICP record deleted"})
}

// handleImportICPRecords 处理批量导入 ICP 记录请求
func (s *MockICPServer) handleImportICPRecords(c *gin.Context) {
	var records []model.ICPRecord
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 导入记录
	s.icpRecordsMutex.Lock()
	for _, record := range records {
		// 规范化域名
		record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))
		s.icpRecords[record.Domain] = record
	}
	s.icpRecordsMutex.Unlock()

	logger.RunLogger.Info().
		Int("count", len(records)).
		Msg("Imported ICP records")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Imported %d ICP records", len(records)),
		"count":   len(records),
	})
}

// handleClearICPRecords 处理清空所有 ICP 记录请求
func (s *MockICPServer) handleClearICPRecords(c *gin.Context) {
	s.icpRecordsMutex.Lock()
	count := len(s.icpRecords)
	s.icpRecords = make(map[string]model.ICPRecord)
	s.icpRecordsMutex.Unlock()

	logger.RunLogger.Info().
		Int("count", count).
		Msg("Cleared all ICP records")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Cleared %d ICP records", count),
		"count":   count,
	})
}

// LoadDefaultICPRecords 加载默认 ICP 记录
func (s *MockICPServer) LoadDefaultICPRecords() error {
	// 默认 ICP 记录
	defaultRecords := []model.ICPRecord{
		{
			Domain:       "example.com",
			ICPNumber:    "京ICP备12345678号-1",
			Owner:        "示例公司",
			Type:         "企业",
			Status:       "已备案",
			ApprovalDate: "2023-01-01",
		},
		{
			Domain:       "example.org",
			ICPNumber:    "京ICP备12345679号-1",
			Owner:        "示例组织",
			Type:         "组织",
			Status:       "已备案",
			ApprovalDate: "2023-01-02",
		},
		{
			Domain:       "example.net",
			ICPNumber:    "京ICP备12345680号-1",
			Owner:        "示例网络",
			Type:         "企业",
			Status:       "已备案",
			ApprovalDate: "2023-01-03",
		},
	}

	// 导入记录
	s.icpRecordsMutex.Lock()
	for _, record := range defaultRecords {
		s.icpRecords[record.Domain] = record
	}
	s.icpRecordsMutex.Unlock()

	logger.RunLogger.Info().
		Int("count", len(defaultRecords)).
		Msg("Loaded default ICP records")

	return nil
}

// LoadICPRecordsFromFile 从文件加载 ICP 记录
func (s *MockICPServer) LoadICPRecordsFromFile(path string) error {
	// 从文件加载
	data, err := loadFromFile(path)
	if err != nil {
		return fmt.Errorf("load from file: %w", err)
	}

	// 反序列化记录
	var records []model.ICPRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("unmarshal records: %w", err)
	}

	// 导入记录
	s.icpRecordsMutex.Lock()
	for _, record := range records {
		// 规范化域名
		record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))
		s.icpRecords[record.Domain] = record
	}
	s.icpRecordsMutex.Unlock()

	logger.RunLogger.Info().
		Int("count", len(records)).
		Str("path", path).
		Msg("Loaded ICP records from file")

	return nil
}

// SaveICPRecordsToFile 保存 ICP 记录到文件
func (s *MockICPServer) SaveICPRecordsToFile(path string) error {
	s.icpRecordsMutex.RLock()
	defer s.icpRecordsMutex.RUnlock()

	// 转换为切片
	records := make([]model.ICPRecord, 0, len(s.icpRecords))
	for _, record := range s.icpRecords {
		records = append(records, record)
	}

	// 序列化记录
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal records: %w", err)
	}

	// 保存到文件
	if err := saveToFile(path, data); err != nil {
		return fmt.Errorf("save to file: %w", err)
	}

	logger.RunLogger.Info().
		Int("count", len(records)).
		Str("path", path).
		Msg("Saved ICP records to file")

	return nil
}
