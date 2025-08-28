package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoggingMiddleware struct {
	Logger *slog.Logger
}

func NewLoggingMiddleware(logger *slog.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		Logger: logger,
	}
}

func (l *LoggingMiddleware) RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		correlationID, _ := c.Get("correlation_id")

		requestLogger := l.Logger.WithGroup("Info").With(
			"request_id", requestID,
			"correlation_id", correlationID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		c.Set("logger", requestLogger)
		c.Next()
	}
}
