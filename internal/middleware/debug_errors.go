package middleware

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/gin-gonic/gin"
)

// DebugErrors stores the configured debug-exposure flag in the request context.
// It is the source of truth for response.HandleError when deciding whether to
// include internal error details in the response envelope.
func DebugErrors(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(messages.CtxDebugErrors, enabled)
		c.Next()
	}
}
