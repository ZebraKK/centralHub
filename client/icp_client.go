package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"centralHub/logger"
	"centralHub/model"
)

/*
	备案查询:
	工信部 https://beian.miit.gov.cn/ 只能网页、小程序
	阿里云
	腾讯云
*/

// ICPClient ICP备案查询客户端接口
type ICPClient interface {
	// QueryICP 查询域名ICP备案信息
	QueryICP(ctx context.Context, domain string) (*model.ICPRecord, error)
}

// RealICPClient 真实ICP备案查询客户端
type RealICPClient struct {
	apiURL    string
	apiKey    string
	apiSecret string
	timeout   time.Duration
	client    *http.Client
}

// NewRealICPClient 创建真实ICP备案查询客户端
func NewRealICPClient(apiURL, apiKey, apiSecret string, timeout time.Duration) *RealICPClient {
	return &RealICPClient{
		apiURL:    apiURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		timeout:   timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// QueryICP 查询域名ICP备案信息
func (c *RealICPClient) QueryICP(ctx context.Context, domain string) (*model.ICPRecord, error) {
	// 记录开始时间
	startTime := time.Now()

	// 构建请求URL
	reqURL := fmt.Sprintf("%s/icp/query?domain=%s", c.apiURL, url.QueryEscape(domain))

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 添加请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-API-Secret", c.apiSecret)

	// 记录请求日志
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("url", reqURL).
		Msg("Querying ICP information")

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Str("url", reqURL).
			Msg("Failed to query ICP information")
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Msg("Failed to read ICP response")
		return nil, fmt.Errorf("read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		logger.RunLogger.Error().
			Int("status_code", resp.StatusCode).
			Str("domain", domain).
			Str("body", string(body)).
			Msg("Unexpected status code from ICP API")
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var icpResp model.ICPResponse
	if err := json.Unmarshal(body, &icpResp); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Str("body", string(body)).
			Msg("Failed to unmarshal ICP response")
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 检查响应码
	if icpResp.Code != 0 {
		logger.RunLogger.Error().
			Int("code", icpResp.Code).
			Str("message", icpResp.Message).
			Str("domain", domain).
			Msg("ICP API error")
		return nil, fmt.Errorf("API error: %s", icpResp.Message)
	}

	// 转换为内部记录
	record := model.ConvertICPDataToRecord(icpResp.Data)

	// 记录响应时间
	duration := time.Since(startTime)
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("icp_number", record.ICPNumber).
		Dur("duration", duration).
		Msg("ICP information queried successfully")

	return &record, nil
}

// MockICPClient 模拟ICP备案查询客户端
type MockICPClient struct {
	mockServerURL string
	timeout       time.Duration
	client        *http.Client
}

// NewMockICPClient 创建模拟ICP备案查询客户端
func NewMockICPClient(mockServerURL string, timeout time.Duration) *MockICPClient {
	return &MockICPClient{
		mockServerURL: mockServerURL,
		timeout:       timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// QueryICP 查询域名ICP备案信息
func (c *MockICPClient) QueryICP(ctx context.Context, domain string) (*model.ICPRecord, error) {
	// 记录开始时间
	startTime := time.Now()

	// 构建请求URL
	reqURL := fmt.Sprintf("%s/icp/query?domain=%s", c.mockServerURL, url.QueryEscape(domain))

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 添加请求头
	req.Header.Set("Content-Type", "application/json")

	// 记录请求日志
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("url", reqURL).
		Msg("Querying mock ICP information")

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Str("url", reqURL).
			Msg("Failed to query mock ICP information")
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Msg("Failed to read mock ICP response")
		return nil, fmt.Errorf("read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		logger.RunLogger.Error().
			Int("status_code", resp.StatusCode).
			Str("domain", domain).
			Str("body", string(body)).
			Msg("Unexpected status code from mock ICP API")
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var record model.ICPRecord
	if err := json.Unmarshal(body, &record); err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("domain", domain).
			Str("body", string(body)).
			Msg("Failed to unmarshal mock ICP response")
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 记录响应时间
	duration := time.Since(startTime)
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("icp_number", record.ICPNumber).
		Dur("duration", duration).
		Msg("Mock ICP information queried successfully")

	return &record, nil
}

// NewICPClient 创建ICP备案查询客户端
// 根据配置决定使用真实客户端还是模拟客户端
func NewICPClient(useMock bool, config map[string]string, timeout time.Duration) ICPClient {
	if useMock {
		mockServerURL := config["mock_server_url"]
		if mockServerURL == "" {
			mockServerURL = "http://localhost:8092"
		}
		return NewMockICPClient(mockServerURL, timeout)
	}

	apiURL := config["api_url"]
	apiKey := config["api_key"]
	apiSecret := config["api_secret"]
	return NewRealICPClient(apiURL, apiKey, apiSecret, timeout)
}
