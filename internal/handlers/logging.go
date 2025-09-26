package handlers

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func GetComponentLogger(c *gin.Context, component string) (*slog.Logger, error) {
	loggerValue, exists := c.Get("logger")
	if !exists {
		return nil, fmt.Errorf("logger not found in context")
	}

	logger := loggerValue.(*slog.Logger)
	return logger.With("component", component), nil
}
func logRequest(logger *slog.Logger, operation string, fields ...interface{}) {
	logger.Info(operation, fields...)
}

func LogError(logger *slog.Logger, operation string, err error, fields ...interface{}) {
	logger.Error(operation, "error", err)
}
