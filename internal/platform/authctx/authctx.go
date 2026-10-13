package authctx

import (
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

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
	CompanySlug string
	IsRoot      bool
	IsActive    bool
	Permissions []string
}

// HasTenantRole matches only a trusted role slug under an actor-owned tenant scope.
// Display names and legacy Role values never establish this authority.
func (u *User) HasTenantRole(tc tenant.TenantContext, slug string) bool {
	return u != nil && u.RoleSlug == slug && slug != "" && !tc.IsRootScope && tc.CompanyID != 0 && u.CompanyID != nil && *u.CompanyID == tc.CompanyID
}

// HasCompany reports whether the authenticated user is assigned to a company.
func (u *User) HasCompany() bool {
	return u != nil && u.CompanyID != nil
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
