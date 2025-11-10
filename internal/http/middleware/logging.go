package middleware

import (
	"auxstream/internal/logger"
	"auxstream/internal/metrics"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		metrics.IncActiveConnections()
		defer metrics.DecActiveConnections()

		c.Next()

		duration := time.Since(startTime)
		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)

		metrics.RecordHTTPRequest(method, path, statusStr, duration.Seconds())

		logLevel := getLogLevel(status)
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		switch logLevel {
		case "debug":
			logger.Debug("HTTP request", fields...)
		case "info":
			logger.Info("HTTP request", fields...)
		case "warn":
			logger.Warn("HTTP request", fields...)
		case "error":
			logger.Error("HTTP request", fields...)
		}
	}
}

func getLogLevel(status int) string {
	switch {
	case status >= 500:
		return "error"
	case status >= 400:
		return "warn"
	case status >= 300:
		return "info"
	default:
		return "debug"
	}
}
