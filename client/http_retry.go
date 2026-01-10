package client

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries     int                              // 最大重试次数
	InitialBackoff time.Duration                    // 初始退避时间
	MaxBackoff     time.Duration                    // 最大退避时间
	BackoffFactor  float64                          // 退避因子
	RetryableFunc  func(*http.Response, error) bool // 判断是否需要重试
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		BackoffFactor:  2.0,
		RetryableFunc:  DefaultRetryableFunc,
	}
}

// DefaultRetryableFunc 默认的重试判断函数
// 对于网络错误或 5xx 状态码进行重试
func DefaultRetryableFunc(resp *http.Response, err error) bool {
	// 网络错误需要重试
	if err != nil {
		return true
	}

	// 5xx 服务器错误需要重试
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}

	// 429 Too Many Requests 需要重试
	if resp != nil && resp.StatusCode == 429 {
		return true
	}

	return false
}

// WithRetry 为 HTTP 客户端添加重试功能
func WithRetry(config *RetryConfig) HTTPClientOption {
	return func(c *HTTPClient) {
		originalClient := c.client
		c.client = &http.Client{
			Timeout:   originalClient.Timeout,
			Transport: &retryTransport{transport: originalClient.Transport, config: config},
		}
	}
}

// retryTransport 实现带重试的 Transport
type retryTransport struct {
	transport http.RoundTripper
	config    *RetryConfig
}

// RoundTrip 实现 http.RoundTripper 接口
func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= rt.config.MaxRetries; attempt++ {
		// 发送请求
		resp, err = rt.getTransport().RoundTrip(req)

		// 判断是否需要重试
		if !rt.config.RetryableFunc(resp, err) {
			return resp, err
		}

		// 最后一次尝试不再等待
		if attempt == rt.config.MaxRetries {
			break
		}

		// 计算退避时间
		backoff := rt.calculateBackoff(attempt)

		// 等待退避时间
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(backoff):
		}
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", rt.config.MaxRetries, err)
	}

	return resp, nil
}

// getTransport 获取底层 Transport
func (rt *retryTransport) getTransport() http.RoundTripper {
	if rt.transport != nil {
		return rt.transport
	}
	return http.DefaultTransport
}

// calculateBackoff 计算退避时间（指数退避）
func (rt *retryTransport) calculateBackoff(attempt int) time.Duration {
	backoff := float64(rt.config.InitialBackoff) * math.Pow(rt.config.BackoffFactor, float64(attempt))
	if backoff > float64(rt.config.MaxBackoff) {
		backoff = float64(rt.config.MaxBackoff)
	}
	return time.Duration(backoff)
}

// DoWithRetry 执行带重试的请求
func (c *HTTPClient) DoWithRetry(ctx context.Context, method, url string, body interface{}, headers map[string]string, retryConfig *RetryConfig) (*http.Response, error) {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		resp, err = c.Request(ctx, method, url, body, headers)

		// 判断是否需要重试
		if !retryConfig.RetryableFunc(resp, err) {
			return resp, err
		}

		// 最后一次尝试不再等待
		if attempt == retryConfig.MaxRetries {
			break
		}

		// 计算退避时间
		backoff := calculateBackoffTime(retryConfig, attempt)

		// 等待退避时间
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", retryConfig.MaxRetries, err)
	}

	return resp, nil
}

// calculateBackoffTime 计算退避时间
func calculateBackoffTime(config *RetryConfig, attempt int) time.Duration {
	backoff := float64(config.InitialBackoff) * math.Pow(config.BackoffFactor, float64(attempt))
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}
	return time.Duration(backoff)
}
