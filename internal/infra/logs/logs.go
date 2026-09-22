package logs

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

func GinMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		switch {
		case status >= http.StatusInternalServerError:
			logger.Error("request", attrs...)
		case status >= http.StatusBadRequest:
			logger.Warn("request", attrs...)
		default:
			logger.Info("request", attrs...)
		}
	}
}
