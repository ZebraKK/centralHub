package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"centralHub/logger"
)

// DNSRecord 表示DNS记录
type DNSRecord struct {
	ID         string `json:"id"`          // 记录ID
	Domain     string `json:"domain"`      // 域名
	RecordType string `json:"record_type"` // 记录类型 (A, CNAME, TXT, etc.)
	Value      string `json:"value"`       // 记录值
	TTL        int    `json:"ttl"`         // TTL (秒)
	CreatedAt  int64  `json:"created_at"`  // 创建时间
	UpdatedAt  int64  `json:"updated_at"`  // 更新时间
}

// DNSClientInterface DNS客户端接口
type DNSClientInterface interface {
	AddRecord(domain, recordType, value string, ttl int) (string, error)
	GetRecords(domain string) ([]DNSRecord, error)
	GetRecordsByType(domain, recordType string) ([]DNSRecord, error)
	DeleteRecord(domain, recordID string) error
	VerifyTXTRecord(domain, expectedValue string) (bool, error)
	VerifyCNAMERecord(domain, expectedTarget string) (bool, error)
}

// RealDNSClient 真实DNS客户端实现
type RealDNSClient struct {
	resolver *net.Resolver
	timeout  time.Duration
	apiURL   string
	apiKey   string
}

// NewRealDNSClient 创建新的真实DNS客户端
func NewRealDNSClient(timeout time.Duration, apiURL, apiKey string) *RealDNSClient {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &RealDNSClient{
		resolver: net.DefaultResolver,
		timeout:  timeout,
		apiURL:   apiURL,
		apiKey:   apiKey,
	}
}

// AddRecord 添加DNS记录
func (c *RealDNSClient) AddRecord(domain, recordType, value string, ttl int) (string, error) {
	// 实际实现中，这里应该调用DNS服务商的API添加记录
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("type", recordType).
		Str("value", value).
		Int("ttl", ttl).
		Msg("Adding DNS record")

	// 这里是调用真实API的代码
	// ...

	// 返回记录ID
	return "mock-record-id", nil
}

// GetRecords 获取所有DNS记录
func (c *RealDNSClient) GetRecords(domain string) ([]DNSRecord, error) {
	// 实际实现中，这里应该调用DNS服务商的API获取记录
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Getting DNS records")

	// 这里是调用真实API的代码
	// ...

	// 返回空记录列表
	return []DNSRecord{}, nil
}

// GetRecordsByType 获取指定类型的DNS记录
func (c *RealDNSClient) GetRecordsByType(domain, recordType string) ([]DNSRecord, error) {
	// 获取所有记录
	records, err := c.GetRecords(domain)
	if err != nil {
		return nil, err
	}

	// 过滤指定类型的记录
	filteredRecords := make([]DNSRecord, 0)
	for _, record := range records {
		if record.RecordType == recordType {
			filteredRecords = append(filteredRecords, record)
		}
	}

	return filteredRecords, nil
}

// DeleteRecord 删除DNS记录
func (c *RealDNSClient) DeleteRecord(domain, recordID string) error {
	// 实际实现中，这里应该调用DNS服务商的API删除记录
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("id", recordID).
		Msg("Deleting DNS record")

	// 这里是调用真实API的代码
	// ...

	return nil
}

// VerifyTXTRecord 验证TXT记录
func (c *RealDNSClient) VerifyTXTRecord(domain, expectedValue string) (bool, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// 查询TXT记录
	records, err := c.resolver.LookupTXT(ctx, domain)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			if dnsErr.IsNotFound {
				return false, nil // 记录不存在
			}
		}
		return false, fmt.Errorf("DNS lookup error: %w", err)
	}

	// 检查是否存在匹配的记录
	for _, record := range records {
		if strings.TrimSpace(record) == expectedValue {
			return true, nil
		}
	}

	return false, nil
}

// VerifyCNAMERecord 验证CNAME记录
func (c *RealDNSClient) VerifyCNAMERecord(domain, expectedTarget string) (bool, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// 查询CNAME记录
	cname, err := c.resolver.LookupCNAME(ctx, domain)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			if dnsErr.IsNotFound {
				return false, nil // 记录不存在
			}
		}
		return false, fmt.Errorf("DNS lookup error: %w", err)
	}

	// 去除末尾的点并比较
	cname = strings.TrimSuffix(cname, ".")
	expectedTarget = strings.TrimSuffix(expectedTarget, ".")

	return cname == expectedTarget, nil
}

// MockDNSClient Mock DNS客户端实现
type MockDNSClient struct {
	mockServerURL string
}

// NewMockDNSClient 创建新的Mock DNS客户端
func NewMockDNSClient(mockServerURL string) *MockDNSClient {
	if mockServerURL == "" {
		mockServerURL = "http://localhost:8090"
	}

	return &MockDNSClient{
		mockServerURL: mockServerURL,
	}
}

// AddRecord 添加DNS记录
func (c *MockDNSClient) AddRecord(domain, recordType, value string, ttl int) (string, error) {
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("type", recordType).
		Str("value", value).
		Int("ttl", ttl).
		Msg("Adding DNS record (mock)")

	// 创建请求记录
	record := DNSRecord{
		Domain:     domain,
		RecordType: recordType,
		Value:      value,
		TTL:        ttl,
	}

	// 调用Mock服务添加记录
	url := fmt.Sprintf("%s/dns/record", c.mockServerURL)
	resp, err := postJSON(url, record)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to add record: status code %d", resp.StatusCode)
	}

	// 解析响应
	var result DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.ID, nil
}

// GetRecords 获取所有DNS记录
func (c *MockDNSClient) GetRecords(domain string) ([]DNSRecord, error) {
	logger.RunLogger.Info().
		Str("domain", domain).
		Msg("Getting DNS records (mock)")

	// 调用Mock服务获取记录
	url := fmt.Sprintf("%s/dns/records/%s", c.mockServerURL, domain)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get records: status code %d", resp.StatusCode)
	}

	// 解析响应
	var records []DNSRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return records, nil
}

// GetRecordsByType 获取指定类型的DNS记录
func (c *MockDNSClient) GetRecordsByType(domain, recordType string) ([]DNSRecord, error) {
	// 获取所有记录
	records, err := c.GetRecords(domain)
	if err != nil {
		return nil, err
	}

	// 过滤指定类型的记录
	filteredRecords := make([]DNSRecord, 0)
	for _, record := range records {
		if record.RecordType == recordType {
			filteredRecords = append(filteredRecords, record)
		}
	}

	return filteredRecords, nil
}

// DeleteRecord 删除DNS记录
func (c *MockDNSClient) DeleteRecord(domain, recordID string) error {
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("id", recordID).
		Msg("Deleting DNS record (mock)")

	// 调用Mock服务删除记录
	url := fmt.Sprintf("%s/dns/record/%s/%s", c.mockServerURL, domain, recordID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete record: status code %d", resp.StatusCode)
	}

	return nil
}

// VerifyTXTRecord 验证TXT记录
func (c *MockDNSClient) VerifyTXTRecord(domain, expectedValue string) (bool, error) {
	logger.RunLogger.Info().
		Str("domain", domain).
		Str("value", expectedValue).
		Msg("Verifying TXT record (mock)")

	// 调用Mock服务验证TXT记录
	url := fmt.Sprintf("%s/dns/verify/txt?domain=%s&value=%s", c.mockServerURL, domain, expectedValue)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to verify TXT record: status code %d", resp.StatusCode)
	}

	// 解析响应
	var result struct {
		Verified bool   `json:"verified"`
		Message  string `json:"message,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	return result.Verified, nil
}

// VerifyCNAMERecord 验证CNAME记录
func (c *MockDNSClient) VerifyCNAMERecord(domain, expectedTarget string) (bool, error) {
	// 获取所有记录
	records, err := c.GetRecords(domain)
	if err != nil {
		return false, err
	}

	// 查找CNAME记录并验证
	for _, record := range records {
		if record.RecordType == "CNAME" {
			// 去除末尾的点并比较
			value := strings.TrimSuffix(record.Value, ".")
			target := strings.TrimSuffix(expectedTarget, ".")
			if value == target {
				return true, nil
			}
		}
	}

	return false, nil
}

// 辅助函数，用于发送JSON请求
func postJSON(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
