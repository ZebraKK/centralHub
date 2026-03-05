package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"centralHub/logger"
)

// HTTPClient 封装HTTP请求客户端
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
	timeout time.Duration
}

// HTTPClientOption HTTP客户端选项函数
type HTTPClientOption func(*HTTPClient)

// NewHTTPClient 创建HTTP客户端实例
func NewHTTPClient(options ...HTTPClientOption) *HTTPClient {
	// 创建默认客户端
	client := &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
		timeout: 30 * time.Second,
	}

	// 应用选项
	for _, option := range options {
		option(client)
	}

	return client
}

// WithBaseURL 设置基础URL
func WithBaseURL(baseURL string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.baseURL = baseURL
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) HTTPClientOption {
	return func(c *HTTPClient) {
		c.timeout = timeout
		c.client.Timeout = timeout
	}
}

// WithHeader 设置默认请求头
func WithHeader(key, value string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.headers[key] = value
	}
}

// Request 发送HTTP请求
func (c *HTTPClient) Request(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	// 构建完整URL
	reqURL := path
	if c.baseURL != "" && !strings.HasPrefix(path, "http") {
		reqURL = c.baseURL
		if !strings.HasSuffix(c.baseURL, "/") && !strings.HasPrefix(path, "/") {
			reqURL += "/"
		}
		reqURL += path
	}

	// 准备请求体
	var bodyReader io.Reader
	if body != nil {
		switch b := body.(type) {
		case string:
			bodyReader = strings.NewReader(b)
		case []byte:
			bodyReader = bytes.NewReader(b)
		case io.Reader:
			bodyReader = b
		default:
			// 默认序列化为JSON
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal request body failed: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 设置额外请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 如果是JSON请求体但没有设置Content-Type，则设置
	if body != nil && req.Header.Get("Content-Type") == "" {
		if _, ok := body.(string); !ok {
			if _, ok := body.([]byte); !ok {
				if _, ok := body.(io.Reader); !ok {
					req.Header.Set("Content-Type", "application/json")
				}
			}
		}
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Get 发送GET请求
func (c *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.Request(ctx, http.MethodGet, path, nil, headers)
}

// Post 发送POST请求
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, path, body, headers)
}

// Put 发送PUT请求
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.Request(ctx, http.MethodPut, path, body, headers)
}

// Delete 发送DELETE请求
func (c *HTTPClient) Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, path, nil, headers)
}

// GetJSON 发送GET请求并解析JSON响应
func (c *HTTPClient) GetJSON(ctx context.Context, path string, result interface{}, headers map[string]string) error {
	return c.RequestJSON(ctx, http.MethodGet, path, nil, result, headers)
}

// PostJSON 发送POST请求并解析JSON响应
func (c *HTTPClient) PostJSON(ctx context.Context, path string, body, result interface{}, headers map[string]string) error {
	return c.RequestJSON(ctx, http.MethodPost, path, body, result, headers)
}

// PutJSON 发送PUT请求并解析JSON响应
func (c *HTTPClient) PutJSON(ctx context.Context, path string, body, result interface{}, headers map[string]string) error {
	return c.RequestJSON(ctx, http.MethodPut, path, body, result, headers)
}

// DeleteJSON 发送DELETE请求并解析JSON响应
func (c *HTTPClient) DeleteJSON(ctx context.Context, path string, result interface{}, headers map[string]string) error {
	return c.RequestJSON(ctx, http.MethodDelete, path, nil, result, headers)
}

// RequestJSON 发送请求并解析JSON响应
func (c *HTTPClient) RequestJSON(ctx context.Context, method, path string, body, result interface{}, headers map[string]string) error {
	// 确保设置了Accept头
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Accept"]; !ok {
		headers["Accept"] = "application/json"
	}

	// 发送请求
	resp, err := c.Request(ctx, method, path, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response failed: %w", err)
		}
	}

	return nil
}

// GetContent 获取指定URL的内容
func (c *HTTPClient) GetContent(ctx context.Context, urlStr string) (string, error) {
	// 发送GET请求
	headers := map[string]string{
		"Accept": "text/plain",
	}
	resp, err := c.Get(ctx, urlStr, headers)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("url", urlStr).
			Msg("HTTP request failed")
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		logger.RunLogger.Error().
			Int("status_code", resp.StatusCode).
			Str("url", urlStr).
			Msg("HTTP request returned non-OK status")
		return "", fmt.Errorf("request returned status %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.RunLogger.Error().
			Err(err).
			Str("url", urlStr).
			Msg("Failed to read HTTP response body")
		return "", fmt.Errorf("read response failed: %w", err)
	}

	logger.RunLogger.Debug().
		Str("url", urlStr).
		Int("content_length", len(body)).
		Msg("HTTP request completed successfully")

	return string(body), nil
}

// GetContentWithRetry 带重试的获取URL内容
func (c *HTTPClient) GetContentWithRetry(ctx context.Context, urlStr string, maxRetries int, retryInterval time.Duration) (string, error) {
	var lastErr error

	// 重试循环
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 如果不是第一次尝试，等待指定的重试间隔
		if attempt > 0 {
			logger.RunLogger.Info().
				Int("attempt", attempt).
				Int("max_retries", maxRetries).
				Str("url", urlStr).
				Msg("Retrying HTTP request")

			select {
			case <-time.After(retryInterval):
				// 继续重试
			case <-ctx.Done():
				// 上下文取消，停止重试
				return "", fmt.Errorf("context canceled during retry: %w", ctx.Err())
			}
		}

		// 尝试获取内容
		content, err := c.GetContent(ctx, urlStr)
		if err == nil {
			// 成功获取内容，返回
			if attempt > 0 {
				logger.RunLogger.Info().
					Int("attempt", attempt+1).
					Str("url", urlStr).
					Msg("HTTP request succeeded after retries")
			}
			return content, nil
		}

		// 记录错误并继续重试
		lastErr = err
	}

	// 所有重试都失败
	logger.RunLogger.Error().
		Err(lastErr).
		Int("max_retries", maxRetries).
		Str("url", urlStr).
		Msg("All HTTP request attempts failed")

	return "", fmt.Errorf("all %d attempts failed, last error: %w", maxRetries+1, lastErr)
}

// VerifyFileContent 验证URL内容是否包含预期值
func (c *HTTPClient) VerifyFileContent(ctx context.Context, urlStr string, expectedContent string) (bool, error) {
	// 获取URL内容
	content, err := c.GetContentWithRetry(ctx, urlStr, 3, 2*time.Second)
	if err != nil {
		return false, err
	}

	// 验证内容是否匹配
	if content == expectedContent {
		logger.RunLogger.Info().
			Str("url", urlStr).
			Msg("File content verification succeeded")
		return true, nil
	}

	logger.RunLogger.Info().
		Str("url", urlStr).
		Msg("File content verification failed: content does not match")

	return false, nil
}

// BuildURL 构建完整URL
func (c *HTTPClient) BuildURL(path string, queryParams map[string]string) (string, error) {
	// 构建基础URL
	baseURL := path
	if c.baseURL != "" && !strings.HasPrefix(path, "http") {
		baseURL = c.baseURL
		if !strings.HasSuffix(c.baseURL, "/") && !strings.HasPrefix(path, "/") {
			baseURL += "/"
		}
		baseURL += path
	}

	// 解析URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse URL failed: %w", err)
	}

	// 添加查询参数
	if len(queryParams) > 0 {
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}
