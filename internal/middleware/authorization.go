package middleware

import (
	"log/slog"

	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

const rootPermissionMarker = "*"

// PermissionResolver resolves permission slugs for an authenticated user.
type PermissionResolver interface {
	Resolve(publicID string) ([]string, error)
}

// AttachPermissions enriches the authenticated user context with permission slugs.
func AttachPermissions(resolver PermissionResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authctx.FromGin(c)
		if !ok || user == nil {
			c.Next()
			return
		}

		if isRoot(user) {
			user.Permissions = []string{rootPermissionMarker}
			authctx.SetGin(c, user)
			c.Next()
			return
		}

		if resolver == nil {
			user.Permissions = []string{}
			authctx.SetGin(c, user)
			c.Next()
			return
		}

		permissions, err := resolver.Resolve(user.PublicID)
		if err != nil {
			slog.Warn("permission resolution failed", "public_id", user.PublicID, "error", err)
			permissions = []string{}
		}
		user.Permissions = permissions
		authctx.SetGin(c, user)
		c.Next()
	}
}

// RequirePermission allows requests whose authenticated user has the required permission.
func RequirePermission(slug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authctx.FromGin(c)
		if !ok || user == nil {
			response.Unauthorized(c, messages.MsgUnauthorized)
			c.Abort()
			return
		}
		if isRoot(user) || hasPermission(user.Permissions, slug) {
			c.Next()
			return
		}
		response.Forbidden(c, messages.MsgForbidden)
		c.Abort()
	}
}

// RequireRole allows requests whose authenticated user's role slug matches the required role.
func RequireRole(slug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authctx.FromGin(c)
		if !ok || user == nil {
			response.Unauthorized(c, messages.MsgUnauthorized)
			c.Abort()
			return
		}
		if roleSlug(user) == slug {
			c.Next()
			return
		}
		response.Forbidden(c, messages.MsgForbidden)
		c.Abort()
	}
}

func isRoot(user *authctx.User) bool {
	return roleSlug(user) == "root"
}

func roleSlug(user *authctx.User) string {
	if user.RoleSlug != "" {
		return user.RoleSlug
	}
	return user.Role
}

func hasPermission(permissions []string, slug string) bool {
	for _, permission := range permissions {
		if permission == rootPermissionMarker || permission == slug {
			return true
		}
	}
	return false
}
