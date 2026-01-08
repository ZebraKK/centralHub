package model

import "time"

// Response 通用响应结构体
type Response struct {
	Code    int         `json:"code"`               // 响应码
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// PageResponse 分页响应结构体
type PageResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   int64       `json:"total"` // 总数
	Page    int         `json:"page"`  // 当前页
	Size    int         `json:"size"`  // 每页大小
}

// DomainResponse 域名响应结构体
type DomainResponse struct {
	ID         string    `json:"id"`
	DomainName string    `json:"domain_name"`
	Status     string    `json:"status"`
	Owner      string    `json:"owner"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

// CreateTaskResponse 创建任务响应结构体
type CreateTaskResponse struct {
	TaskID  string `json:"task_id"`
	Domain  string `json:"domain"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// 响应码常量
const (
	CodeSuccess      = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// NewPageResponse 创建分页响应
func NewPageResponse(data interface{}, total int64, page, size int) *PageResponse {
	return &PageResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	}
}
