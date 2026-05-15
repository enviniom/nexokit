package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// Logger logs HTTP requests with structured fields.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		rid, _ := c.Get("request_id")
		ridStr := ""
		if rid != nil {
			ridStr = rid.(string)
		}

		log.Info("http request",
			slog.String("request_id", ridStr),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("body_size", c.Writer.Size()),
		)
	}
}
