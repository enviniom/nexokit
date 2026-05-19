package authctx

import "github.com/gin-gonic/gin"

const contextKey = "auth_user"

// User is the sanitized authenticated user injected into request context.
type User struct {
	ID          uint
	PublicID    string
	Email       string
	Name        string
	Role        string
	RoleSlug    string
	RoleID      uint
	CompanyID   *uint
	IsActive    bool
	Permissions []string
}

// SetGin stores the authenticated user in the Gin context.
func SetGin(c *gin.Context, user *User) {
	c.Set(contextKey, user)
}

// FromGin retrieves the authenticated user from the Gin context.
func FromGin(c *gin.Context) (*User, bool) {
	value, exists := c.Get(contextKey)
	if !exists {
		return nil, false
	}
	user, ok := value.(*User)
	return user, ok
}

// PublicIDFromGin returns the authenticated user's public ID when present.
func PublicIDFromGin(c *gin.Context) string {
	user, ok := FromGin(c)
	if !ok || user == nil {
		return ""
	}
	return user.PublicID
}
