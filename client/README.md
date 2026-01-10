# HTTP Client 使用文档

centralHub 项目的 HTTP 客户端封装，提供便捷的 HTTP 请求功能。

## 功能特性

- ✅ 支持所有常用 HTTP 方法（GET、POST、PUT、DELETE、PATCH）
- ✅ 自动 JSON 序列化和反序列化
- ✅ 灵活的配置选项（基础 URL、超时、请求头等）
- ✅ 内置重试机制（指数退避算法）
- ✅ Context 支持（超时控制、取消请求）
- ✅ 自定义 Transport
- ✅ 易于使用的 API

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "yourproject/client"
)

func main() {
    // 创建 HTTP 客户端
    httpClient := client.NewHTTPClient()

    // 发送 GET 请求
    ctx := context.Background()
    resp, err := httpClient.Get(ctx, "https://api.example.com/users", nil)
    if err != nil {
        fmt.Printf("Request failed: %v\n", err)
        return
    }
    defer resp.Body.Close()

    fmt.Printf("Status: %d\n", resp.StatusCode)
}
```

### 使用配置选项

```go
// 创建带配置的客户端
httpClient := client.NewHTTPClient(
    client.WithBaseURL("https://api.example.com"),
    client.WithTimeout(30*time.Second),
    client.WithHeader("Authorization", "Bearer your-token"),
    client.WithHeader("User-Agent", "CentralHub/1.0"),
)

// 发送请求时会自动拼接基础 URL
resp, err := httpClient.Get(ctx, "/users", nil)
```

### JSON 请求和响应

```go
// 定义请求体
requestBody := map[string]interface{}{
    "name":  "张三",
    "email": "<EMAIL_ADDRESS>",
}

// 定义响应体
var responseData map[string]interface{}

// 发送 POST 请求并自动解析 JSON 响应
err := httpClient.PostJSON(ctx, "/users", requestBody, &responseData, nil)
if err != nil {
    fmt.Printf("Request failed: %v\n", err)
    return
}

fmt.Printf("Created user: %+v\n", responseData)
```

### 使用日志记录

```go
import (
    "context"
    "yourproject/client"
    "yourproject/logger"
)

// 获取带上下文的 logger（例如带 reqid）
reqLogger := logger.WithReqID("request-123")

// 使用带 logger 的方法发送请求
var result map[string]interface{}
err := httpClient.GetJSONWithLogger(ctx, "/users", &result, nil, &reqLogger)
if err != nil {
    reqLogger.Error().Err(err).Msg("Failed to get users")
    return
}

// 日志会自动记录请求详情，包括 reqid
reqLogger.Info().Msg("Successfully retrieved users")
```

## API 参考

### 创建客户端

#### NewHTTPClient

```go
func NewHTTPClient(options ...HTTPClientOption) *HTTPClient
```

创建一个新的 HTTP 客户端实例。

**参数:**
- `options`: 可选的配置选项

**返回:**
- `*HTTPClient`: HTTP 客户端实例

### 配置选项

#### WithBaseURL

```go
func WithBaseURL(baseURL string) HTTPClientOption
```

设置基础 URL，所有请求的 URL 都会拼接在基础 URL 后面。

#### WithTimeout

```go
func WithTimeout(timeout time.Duration) HTTPClientOption
```

设置请求超时时间（默认 30 秒）。

#### WithHeader

```go
func WithHeader(key, value string) HTTPClientOption
```

设置默认请求头，这些请求头会被添加到所有请求中。

#### WithTransport

```go
func WithTransport(transport *http.Transport) HTTPClientOption
```

自定义 HTTP Transport。

#### WithRetry

```go
func WithRetry(config *RetryConfig) HTTPClientOption
```

启用重试机制。

### HTTP 方法

#### Get

```go
func (c *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error)
```

发送 GET 请求。

#### Post

```go
func (c *HTTPClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error)
```

发送 POST 请求，自动设置 `Content-Type: application/json`。

#### Put

```go
func (c *HTTPClient) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error)
```

发送 PUT 请求。

#### Delete

```go
func (c *HTTPClient) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error)
```

发送 DELETE 请求。

#### Patch

```go
func (c *HTTPClient) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*http.Response, error)
```

发送 PATCH 请求。

### JSON 便捷方法

#### GetJSON

```go
func (c *HTTPClient) GetJSON(ctx context.Context, url string, result interface{}, headers map[string]string) error
```

发送 GET 请求并自动解析 JSON 响应到 result。

#### GetJSONWithLogger

```go
func (c *HTTPClient) GetJSONWithLogger(ctx context.Context, url string, result interface{}, headers map[string]string, logger *zerolog.Logger) error
```

发送 GET 请求并自动解析 JSON 响应（带日志记录）。

#### PostJSON

```go
func (c *HTTPClient) PostJSON(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string) error
```

发送 POST 请求并自动解析 JSON 响应。

#### PostJSONWithLogger

```go
func (c *HTTPClient) PostJSONWithLogger(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string, logger *zerolog.Logger) error
```

发送 POST 请求并自动解析 JSON 响应（带日志记录）。

#### PutJSON

```go
func (c *HTTPClient) PutJSON(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string) error
```

发送 PUT 请求并自动解析 JSON 响应。

#### PutJSONWithLogger

```go
func (c *HTTPClient) PutJSONWithLogger(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string, logger *zerolog.Logger) error
```

发送 PUT 请求并自动解析 JSON 响应（带日志记录）。

#### DeleteJSON

```go
func (c *HTTPClient) DeleteJSON(ctx context.Context, url string, result interface{}, headers map[string]string) error
```

发送 DELETE 请求并自动解析 JSON 响应。

#### DeleteJSONWithLogger

```go
func (c *HTTPClient) DeleteJSONWithLogger(ctx context.Context, url string, result interface{}, headers map[string]string, logger *zerolog.Logger) error
```

发送 DELETE 请求并自动解析 JSON 响应（带日志记录）。

## 重试机制

### 使用默认重试配置

```go
// 创建带重试的客户端
httpClient := client.NewHTTPClient(
    client.WithBaseURL("https://api.example.com"),
    client.WithRetry(client.DefaultRetryConfig()),
)

// 请求失败时会自动重试（最多 3 次）
resp, err := httpClient.Get(ctx, "/unstable-endpoint", nil)
```

### 自定义重试配置

```go
// 自定义重试策略
retryConfig := &client.RetryConfig{
    MaxRetries:     5,                        // 最大重试次数
    InitialBackoff: 200 * time.Millisecond,   // 初始退避时间
    MaxBackoff:     10 * time.Second,         // 最大退避时间
    BackoffFactor:  2.0,                      // 退避因子（指数退避）
    RetryableFunc: func(resp *http.Response, err error) bool {
        // 自定义重试逻辑
        if err != nil {
            return true // 网络错误重试
        }
        // 只对 503 和 429 状态码重试
        return resp != nil && (resp.StatusCode == 503 || resp.StatusCode == 429)
    },
}

httpClient := client.NewHTTPClient(
    client.WithRetry(retryConfig),
)
```

### 默认重试行为

默认的重试配置会在以下情况重试：
- 网络错误
- HTTP 5xx 服务器错误
- HTTP 429 Too Many Requests

重试使用指数退避算法：
- 第 1 次重试: 100ms
- 第 2 次重试: 200ms
- 第 3 次重试: 400ms
- 最多重试 3 次

## 日志记录

HTTP 客户端支持在每个请求中传入 logger，实现带上下文的日志记录：

### 基本日志记录

```go
import (
    "github.com/rs/zerolog/log"
    "yourproject/client"
)

// 使用全局 logger
logger := log.Logger
var result map[string]interface{}

err := httpClient.GetJSONWithLogger(ctx, "/api/users", &result, nil, &logger)
if err != nil {
    logger.Error().Err(err).Msg("Request failed")
    return
}
```

### 使用带上下文的 Logger

```go
import (
    "yourproject/logger"
    "yourproject/client"
)

// 在 HTTP 请求处理中，使用带 reqid 的 logger
func HandleRequest(reqid string) {
    // 创建带 reqid 的 logger
    reqLogger := logger.WithReqID(reqid)
    
    httpClient := client.NewHTTPClient(
        client.WithBaseURL("https://api.example.com"),
    )
    
    // 使用带 logger 的方法
    var users []map[string]interface{}
    err := httpClient.GetJSONWithLogger(ctx, "/users", &users, nil, &reqLogger)
    if err != nil {
        reqLogger.Error().Err(err).Msg("Failed to fetch users")
        return
    }
    
    reqLogger.Info().Int("count", len(users)).Msg("Successfully fetched users")
}
```

### 日志输出示例

使用 logger 后，HTTP 请求会自动记录详细信息：

```json
{
  "level": "debug",
  "reqid": "request-123",
  "method": "GET",
  "url": "https://api.example.com/users",
  "time": "2026-01-09T12:00:00+08:00",
  "msg": "Sending HTTP request"
}

{
  "level": "debug",
  "reqid": "request-123",
  "method": "GET",
  "url": "https://api.example.com/users",
  "status": 200,
  "time": "2026-01-09T12:00:01+08:00",
  "msg": "HTTP request completed"
}
```

### 带 Logger 和不带 Logger 的方法

为了保持向后兼容，HTTP 客户端提供两套方法：

**不带 Logger 的方法**（原有方法）：
- `Get()`, `Post()`, `Put()`, `Delete()`, `Patch()`
- `GetJSON()`, `PostJSON()`, `PutJSON()`, `DeleteJSON()`

**带 Logger 的方法**（新增方法）：
- `GetWithLogger()`, `PostWithLogger()`, `PutWithLogger()`, `DeleteWithLogger()`, `PatchWithLogger()`
- `GetJSONWithLogger()`, `PostJSONWithLogger()`, `PutJSONWithLogger()`, `DeleteJSONWithLogger()`

根据需要选择合适的方法：
- 如果需要日志记录上下文（如 reqid），使用带 Logger 的方法
- 如果不需要日志记录，使用普通方法

## 高级用法

### 使用 Context 控制超时

```go
// 创建带 5 秒超时的 Context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 请求会在 5 秒后超时
resp, err := httpClient.Get(ctx, "/slow-endpoint", nil)
if err != nil {
    fmt.Printf("Request timed out: %v\n", err)
}
```

### 自定义请求头

```go
// 为单个请求添加自定义请求头
headers := map[string]string{
    "X-Request-ID": "unique-id-12345",
    "X-Custom-Header": "custom-value",
}

resp, err := httpClient.Get(ctx, "/users", headers)
```

### 手动处理响应

```go
resp, err := httpClient.Post(ctx, "/users", requestBody, nil)
if err != nil {
    return err
}
defer resp.Body.Close()

// 读取响应体
body, err := io.ReadAll(resp.Body)
if err != nil {
    return err
}

// 检查状态码
if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

// 自定义解析逻辑
var result MyCustomType
if err := json.Unmarshal(body, &result); err != nil {
    return err
}
```

## 错误处理

HTTP 客户端会返回详细的错误信息：

```go
resp, err := httpClient.Get(ctx, "/api/endpoint", nil)
if err != nil {
    // 可能的错误类型：
    // - 网络错误
    // - 超时错误
    // - Context 取消错误
    fmt.Printf("Request failed: %v\n", err)
    return
}

// 检查 HTTP 状态码
if resp.StatusCode >= 400 {
    fmt.Printf("HTTP error: %d\n", resp.StatusCode)
}
```

## 最佳实践

1. **使用 Context**: 始终传递 Context 以便控制请求的生命周期
2. **关闭响应体**: 记得 `defer resp.Body.Close()`
3. **复用客户端**: 创建一个客户端实例并复用，避免频繁创建
4. **设置合理的超时**: 根据业务需求设置适当的超时时间
5. **使用 JSON 便捷方法**: 对于 JSON API，使用 `GetJSON`、`PostJSON` 等方法
6. **启用重试**: 对于不稳定的外部 API，启用重试机制
7. **使用日志记录**: 在生产环境中使用带 Logger 的方法，传递带上下文的 logger（如 reqid），便于追踪和调试

## 示例代码

完整的使用示例请参考 `http_example.go` 文件。

## 注意事项

- 默认超时时间为 30 秒，可以通过 `WithTimeout` 修改
- JSON 请求会自动设置 `Content-Type: application/json`
- 重试机制使用指数退避算法，避免过度请求
- Context 取消会立即中断请求
- 所有方法都是线程安全的

## 许可证

MIT License
