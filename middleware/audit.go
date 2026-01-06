package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	//"github.com/rs/zerolog"

	"centralHub/logger"
)

// AuditLog zerolog审计日志中间件
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		// 处理请求
		c.Next()
		// 计算耗时
		latency := time.Since(start)

		// 构建审计日志的结构化字段
		event := logger.AuditLogger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent())

		// 若有请求错误，添加错误信息
		if len(c.Errors) > 0 {
			event = event.Str("errors", c.Errors.String())
		}

		// 写入审计日志（Info级别）
		event.Msg("HTTP Request Audit")
	}
}

// AuditLogWithReqID 结合审计日志和请求ID中间件
func AuditLogWithReqID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成或获取reqid
		reqID := c.GetHeader("X-Request-Id")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("reqid", reqID)
		c.Writer.Header().Set("X-Request-Id", reqID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		event := logger.AuditLogger.Info().
			Str("reqid", reqID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent())

		if len(c.Errors) > 0 {
			event = event.Str("errors", c.Errors.String())
		}
		event.Msg("HTTP Request Audit")
	}
}
