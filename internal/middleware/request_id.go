package middleware

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID attaches a unique request ID to each incoming request.
// If the client already provided one via the X-Request-ID header, it is preserved.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(messages.HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(messages.CtxRequestID, rid)
		c.Writer.Header().Set(messages.HeaderRequestID, rid)
		c.Next()
	}
}
