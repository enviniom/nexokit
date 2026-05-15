package middleware

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
	"log/slog"
)

// Recovery catches panics, logs them, and returns a structured 500 response.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				rid, _ := c.Get("request_id")
				ridStr, _ := rid.(string)
				log.Error(messages.MsgPanicLog,
					slog.String("request_id", ridStr),
					slog.Any("error", r),
					slog.String("path", c.Request.URL.Path),
				)
				response.InternalServerError(c, messages.MsgPanicRecovered)
				c.Abort()
			}
		}()
		c.Next()
	}
}
