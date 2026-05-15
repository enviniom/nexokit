package middleware

import (
	"net/http"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/gin-gonic/gin"
)

// CORS configures Cross-Origin Resource Sharing headers.
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := cfg.CORS.AllowedOrigins
		if origin == "" {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", messages.CORSAllowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", messages.CORSAllowedHeaders)
		c.Writer.Header().Set("Access-Control-Max-Age", messages.CORSMaxAge)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
