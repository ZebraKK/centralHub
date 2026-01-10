package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// HTTPClient 封装 HTTP 客户端
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// HTTPClientOption 配置选项
type HTTPClientOption func(*HTTPClient)

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient(options ...HTTPClientOption) *HTTPClient {
	client := &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}

	// 应用配置选项
	for _, opt := range options {
		opt(client)
	}

	return client
}

// WithBaseURL 设置基础 URL
func WithBaseURL(baseURL string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.baseURL = baseURL
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) HTTPClientOption {
	return func(c *HTTPClient) {
		c.client.Timeout = timeout
	}
}

// WithHeader 设置默认请求头
func WithHeader(key, value string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.headers[key] = value
	}
}

// WithTransport 自定义 Transport
func WithTransport(transport *http.Transport) HTTPClientOption {
	return func(c *HTTPClient) {
		c.client.Transport = transport
	}
}

// Request 发送 HTTP 请求
func (c *HTTPClient) Request(ctx context.Context, method, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.RequestWithLogger(ctx, method, url, body, headers, nil)
}

// RequestWithLogger 发送 HTTP 请求（带日志记录）
func (c *HTTPClient) RequestWithLogger(ctx context.Context, method, url string, body interface{}, headers map[string]string, logger *zerolog.Logger) (*http.Response, error) {
	var reqBody io.Reader

	// 处理请求体
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			if logger != nil {
				logger.Error().Err(err).Msg("Failed to marshal request body")
			}
			return nil, fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 构建完整 URL
	fullURL := url
	if c.baseURL != "" {
		fullURL = c.baseURL + url
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		if logger != nil {
			logger.Error().Err(err).Str("method", method).Str("url", fullURL).Msg("Failed to create HTTP request")
		}
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 设置自定义请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	if logger != nil {
		logger.Debug().Str("method", method).Str("url", fullURL).Msg("Sending HTTP request")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if logger != nil {
			logger.Error().Err(err).Str("method", method).Str("url", fullURL).Msg("HTTP request failed")
		}
		return nil, fmt.Errorf("do request failed: %w", err)
	}

	if logger != nil {
		logger.Debug().Str("method", method).Str("url", fullURL).Int("status", resp.StatusCode).Msg("HTTP request completed")
	}

	return resp, nil
}

// Get 发送 GET 请求
func (c *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return c.Request(ctx, http.MethodGet, url, nil, headers)
}

// GetWithLogger 发送 GET 请求（带日志记录）
func (c *HTTPClient) GetWithLogger(ctx context.Context, url string, headers map[string]string, logger *zerolog.Logger) (*http.Response, error) {
	return c.RequestWithLogger(ctx, http.MethodGet, url, nil, headers, logger)
}

// Post 发送 POST 请求
func (c *HTTPClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return c.Request(ctx, http.MethodPost, url, body, headers)
}

// PostWithLogger 发送 POST 请求（带日志记录）
func (c *HTTPClient) PostWithLogger(ctx context.Context, url string, body interface{}, headers map[string]string, logger *zerolog.Logger) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return c.RequestWithLogger(ctx, http.MethodPost, url, body, headers, logger)
}

// Put 发送 PUT 请求
func (c *HTTPClient) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return c.Request(ctx, http.MethodPut, url, body, headers)
}

// PutWithLogger 发送 PUT 请求（带日志记录）
func (c *HTTPClient) PutWithLogger(ctx context.Context, url string, body interface{}, headers map[string]string, logger *zerolog.Logger) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return c.RequestWithLogger(ctx, http.MethodPut, url, body, headers, logger)
}

// Delete 发送 DELETE 请求
func (c *HTTPClient) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, url, nil, headers)
}

// DeleteWithLogger 发送 DELETE 请求（带日志记录）
func (c *HTTPClient) DeleteWithLogger(ctx context.Context, url string, headers map[string]string, logger *zerolog.Logger) (*http.Response, error) {
	return c.RequestWithLogger(ctx, http.MethodDelete, url, nil, headers, logger)
}

// Patch 发送 PATCH 请求
func (c *HTTPClient) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return c.Request(ctx, http.MethodPatch, url, body, headers)
}

// PatchWithLogger 发送 PATCH 请求（带日志记录）
func (c *HTTPClient) PatchWithLogger(ctx context.Context, url string, body interface{}, headers map[string]string, logger *zerolog.Logger) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return c.RequestWithLogger(ctx, http.MethodPatch, url, body, headers, logger)
}

// ParseResponse 解析响应体到结构体
func ParseResponse(resp *http.Response, result interface{}) error {
	return ParseResponseWithLogger(resp, result, nil)
}

// ParseResponseWithLogger 解析响应体到结构体（带日志记录）
func ParseResponseWithLogger(resp *http.Response, result interface{}, logger *zerolog.Logger) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if logger != nil {
			logger.Error().Err(err).Msg("Failed to read response body")
		}
		return fmt.Errorf("read response body failed: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if logger != nil {
			logger.Warn().Int("status", resp.StatusCode).Str("body", string(body)).Msg("HTTP request returned error status")
		}
		return fmt.Errorf("http status %d: %s", resp.StatusCode, string(body))
	}

	// 解析 JSON
	if err := json.Unmarshal(body, result); err != nil {
		if logger != nil {
			logger.Error().Err(err).Str("body", string(body)).Msg("Failed to unmarshal response")
		}
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	return nil
}

// GetJSON 发送 GET 请求并解析 JSON 响应
func (c *HTTPClient) GetJSON(ctx context.Context, url string, result interface{}, headers map[string]string) error {
	resp, err := c.Get(ctx, url, headers)
	if err != nil {
		return err
	}
	return ParseResponse(resp, result)
}

// GetJSONWithLogger 发送 GET 请求并解析 JSON 响应（带日志记录）
func (c *HTTPClient) GetJSONWithLogger(ctx context.Context, url string, result interface{}, headers map[string]string, logger *zerolog.Logger) error {
	resp, err := c.GetWithLogger(ctx, url, headers, logger)
	if err != nil {
		return err
	}
	return ParseResponseWithLogger(resp, result, logger)
}

// PostJSON 发送 POST 请求并解析 JSON 响应
func (c *HTTPClient) PostJSON(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string) error {
	resp, err := c.Post(ctx, url, body, headers)
	if err != nil {
		return err
	}
	return ParseResponse(resp, result)
}

// PostJSONWithLogger 发送 POST 请求并解析 JSON 响应（带日志记录）
func (c *HTTPClient) PostJSONWithLogger(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string, logger *zerolog.Logger) error {
	resp, err := c.PostWithLogger(ctx, url, body, headers, logger)
	if err != nil {
		return err
	}
	return ParseResponseWithLogger(resp, result, logger)
}

// PutJSON 发送 PUT 请求并解析 JSON 响应
func (c *HTTPClient) PutJSON(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string) error {
	resp, err := c.Put(ctx, url, body, headers)
	if err != nil {
		return err
	}
	return ParseResponse(resp, result)
}

// PutJSONWithLogger 发送 PUT 请求并解析 JSON 响应（带日志记录）
func (c *HTTPClient) PutJSONWithLogger(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string, logger *zerolog.Logger) error {
	resp, err := c.PutWithLogger(ctx, url, body, headers, logger)
	if err != nil {
		return err
	}
	return ParseResponseWithLogger(resp, result, logger)
}

// DeleteJSON 发送 DELETE 请求并解析 JSON 响应
func (c *HTTPClient) DeleteJSON(ctx context.Context, url string, result interface{}, headers map[string]string) error {
	resp, err := c.Delete(ctx, url, headers)
	if err != nil {
		return err
	}
	return ParseResponse(resp, result)
}

// DeleteJSONWithLogger 发送 DELETE 请求并解析 JSON 响应（带日志记录）
func (c *HTTPClient) DeleteJSONWithLogger(ctx context.Context, url string, result interface{}, headers map[string]string, logger *zerolog.Logger) error {
	resp, err := c.DeleteWithLogger(ctx, url, headers, logger)
	if err != nil {
		return err
	}
	return ParseResponseWithLogger(resp, result, logger)
}
