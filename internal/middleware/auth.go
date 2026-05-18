package middleware

import (
	"net/http"
	"strings"

	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/token"
	"github.com/gin-gonic/gin"
)

// AccessParser validates access tokens.
type AccessParser interface {
	ParseAccess(token string) (*token.AccessClaims, error)
}

// AuthUserLookup resolves sanitized users for authenticated requests.
type AuthUserLookup interface {
	GetAuthUser(publicID string) (*authctx.User, error)
}

// Auth validates Bearer access tokens and injects the active user into context.
func Auth(parser AccessParser, lookup AuthUserLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			response.Unauthorized(c, messages.MsgUnauthorized)
			c.Abort()
			return
		}

		accessToken := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
		if accessToken == "" {
			response.Unauthorized(c, messages.MsgUnauthorized)
			c.Abort()
			return
		}

		claims, err := parser.ParseAccess(accessToken)
		if err != nil || claims == nil || claims.Sub == "" {
			response.Unauthorized(c, messages.MsgUnauthorized)
			c.Abort()
			return
		}

		user, err := lookup.GetAuthUser(claims.Sub)
		if err != nil || user == nil {
			response.Unauthorized(c, messages.MsgUnauthorized)
			c.Abort()
			return
		}
		if !user.IsActive {
			response.Error(c, http.StatusUnauthorized, messages.MsgUnauthorized, nil)
			c.Abort()
			return
		}

		authctx.SetGin(c, user)
		c.Next()
	}
}
