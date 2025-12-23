package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

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
		auditFields := []zerolog.LogObjectMarshaler{
			zerolog.String("method", c.Request.Method),
			zerolog.String("path", c.Request.URL.Path),
			zerolog.Int("status", c.Writer.Status()),
			zerolog.Duration("latency", latency),
			zerolog.String("client_ip", c.ClientIP()),
			zerolog.String("user_agent", c.Request.UserAgent()),
		}
		// 若有请求错误，添加错误信息
		if len(c.Errors) > 0 {
			auditFields = append(auditFields, zerolog.String("errors", c.Errors.String()))
		}

		// 写入审计日志（Info级别）
		logger.AuditLogger.Info().Object("request", zerolog.Objs(auditFields...)).Msg("HTTP Request Audit")
	}
}
