package permissions

import "github.com/enviniom/nexokit/internal/shared"

const (
	ActionIndex             = "index"
	ActionView              = "view"
	ActionList              = "list"
	ActionCreate            = "create"
	ActionUpdate            = "update"
	ActionDelete            = "delete"
	ActionManage            = "manage"
	ActionChangeRole        = "change_role"
	ActionAssignPermissions = "assign_permissions"
	ActionSelect            = "select"
)

// Permission represents an RBAC permission with UI-friendly metadata.
type Permission struct {
	shared.BaseModel
	Slug         string `gorm:"type:varchar(120);uniqueIndex;not null"`
	Name         string `gorm:"type:varchar(120);not null"`
	Module       string `gorm:"type:varchar(80);not null;index"`
	Action       string `gorm:"type:varchar(80);not null"`
	Description  string
	IsSystem     bool `gorm:"not null;default:false"`
	DisplayOrder int  `gorm:"not null;default:0;index"`
}

// RolePermission links roles to permissions.
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey;not null;index"`
	PermissionID uint `gorm:"primaryKey;not null;index"`
}
