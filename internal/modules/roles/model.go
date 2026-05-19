package roles

import (
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/shared"
)

const (
	// RootRoleSlug is the stable business identifier for the root role.
	RootRoleSlug = "root"
	// AdminRoleSlug is the stable business identifier for the admin role.
	AdminRoleSlug = "admin"
	// UserRoleSlug is the stable business identifier for the default user role.
	UserRoleSlug = "user"
)

// Role represents a user role in the system.
type Role struct {
	shared.BaseModel
	Name        string `gorm:"uniqueIndex;not null"`
	Slug        string `gorm:"uniqueIndex;not null"`
	Description string
	IsSystem    bool                     `gorm:"not null;default:false"`
	Permissions []permissions.Permission `gorm:"many2many:role_permissions;joinForeignKey:RoleID;joinReferences:PermissionID"`
}
