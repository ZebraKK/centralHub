package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// 示例：基本使用

// ExampleBasicUsage 基本使用示例
func ExampleBasicUsage() {
	// 创建 HTTP 客户端
	client := NewHTTPClient()

	// 发送 GET 请求
	ctx := context.Background()
	resp, err := client.Get(ctx, "https://api.example.com/users", nil)
	if err != nil {
		fmt.Printf("GET request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
}

// ExampleWithBaseURL 使用基础 URL
func ExampleWithBaseURL() {
	// 创建带基础 URL 的客户端
	client := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
		WithTimeout(10*time.Second),
	)

	ctx := context.Background()

	// 发送请求（会自动拼接基础 URL）
	resp, err := client.Get(ctx, "/users/123", nil)
	if err != nil {
		fmt.Printf("GET request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

// ExampleWithHeaders 使用自定义请求头
func ExampleWithHeaders() {
	// 创建带默认请求头的客户端
	client := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
		WithHeader("Authorization", "Bearer your-token-here"),
		WithHeader("User-Agent", "CentralHub/1.0"),
	)

	ctx := context.Background()

	// 发送请求，可以额外添加请求头
	headers := map[string]string{
		"X-Request-ID": "12345",
	}
	resp, err := client.Get(ctx, "/users", headers)
	if err != nil {
		fmt.Printf("GET request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

// ExamplePostJSON POST JSON 请求示例
func ExamplePostJSON() {
	client := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
	)

	ctx := context.Background()

	// 请求体
	requestBody := map[string]interface{}{
		"name":  "<PERSON>",
		"email": "<EMAIL_ADDRESS>",
	}

	// 响应体
	var responseData map[string]interface{}

	// 发送 POST 请求并解析响应
	err := client.PostJSON(ctx, "/users", requestBody, &responseData, nil)
	if err != nil {
		fmt.Printf("POST request failed: %v\n", err)
		return
	}

	fmt.Printf("Response: %+v\n", responseData)
}

// ExampleWithRetry 使用重试机制
func ExampleWithRetry() {
	// 创建带重试功能的客户端
	retryConfig := DefaultRetryConfig()
	retryConfig.MaxRetries = 5

	client := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
		WithRetry(retryConfig),
	)

	ctx := context.Background()

	// 发送请求（失败时会自动重试）
	resp, err := client.Get(ctx, "/unstable-endpoint", nil)
	if err != nil {
		fmt.Printf("Request failed after retries: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Request succeeded with status: %d\n", resp.StatusCode)
}

// ExampleCustomRetry 自定义重试逻辑
func ExampleCustomRetry() {
	// 自定义重试配置
	retryConfig := &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 200 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		BackoffFactor:  3.0,
		RetryableFunc: func(resp *http.Response, err error) bool {
			// 自定义重试逻辑：只对特定错误重试
			if err != nil {
				return true
			}
			// 对 503 Service Unavailable 重试
			return resp != nil && resp.StatusCode == 503
		},
	}

	client := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
	)

	ctx := context.Background()

	// 使用自定义重试配置
	resp, err := client.DoWithRetry(ctx, "GET", "/service", nil, nil, retryConfig)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

// ExampleCompleteWorkflow 完整的工作流示例
func ExampleCompleteWorkflow() {
	// 创建功能齐全的 HTTP 客户端
	client := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
		WithTimeout(30*time.Second),
		WithHeader("Authorization", "Bearer token"),
		WithRetry(DefaultRetryConfig()),
	)

	ctx := context.Background()

	// 1. GET 请求 - 获取用户列表
	var users []map[string]interface{}
	if err := client.GetJSON(ctx, "/users", &users, nil); err != nil {
		fmt.Printf("Failed to get users: %v\n", err)
		return
	}
	fmt.Printf("Got %d users\n", len(users))

	// 2. POST 请求 - 创建新用户
	newUser := map[string]interface{}{
		"name":  "<PERSON>",
		"email": "<EMAIL_ADDRESS>",
	}
	var createdUser map[string]interface{}
	if err := client.PostJSON(ctx, "/users", newUser, &createdUser, nil); err != nil {
		fmt.Printf("Failed to create user: %v\n", err)
		return
	}
	fmt.Printf("Created user: %+v\n", createdUser)

	// 3. PUT 请求 - 更新用户
	updateData := map[string]interface{}{
		"name": "<PERSON>",
	}
	var updatedUser map[string]interface{}
	if err := client.PutJSON(ctx, "/users/123", updateData, &updatedUser, nil); err != nil {
		fmt.Printf("Failed to update user: %v\n", err)
		return
	}
	fmt.Printf("Updated user: %+v\n", updatedUser)

	// 4. DELETE 请求 - 删除用户
	var deleteResult map[string]interface{}
	if err := client.DeleteJSON(ctx, "/users/123", &deleteResult, nil); err != nil {
		fmt.Printf("Failed to delete user: %v\n", err)
		return
	}
	fmt.Printf("Delete result: %+v\n", deleteResult)
}

// ExampleContextTimeout 使用上下文超时
func ExampleContextTimeout() {
	client := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
	)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 请求会在 5 秒后超时
	resp, err := client.Get(ctx, "/slow-endpoint", nil)
	if err != nil {
		fmt.Printf("Request timed out: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
