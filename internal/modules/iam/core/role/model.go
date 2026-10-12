package role

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"github.com/enviniom/nexokit/internal/shared"
)

// IAMRole is the IAM-local role model used to avoid cross-module imports.
type IAMRole struct {
	shared.BaseModel
	Name        string `gorm:"not null"`
	Slug        string `gorm:"not null"`
	CompanyID   *uint
	Company     IAMCompany `gorm:"foreignKey:CompanyID"`
	Description string
	IsSystem    bool                       `gorm:"not null;default:false"`
	Permissions []permission.IAMPermission `gorm:"many2many:role_permissions;joinForeignKey:RoleID;joinReferences:PermissionID"`
}

func (IAMRole) TableName() string { return "roles" }

// IAMCompany is a local partial model for role preload compatibility.
type IAMCompany struct {
	shared.BaseModelSimple
	Name string `gorm:"not null"`
	Slug string `gorm:"not null"`
}

func (IAMCompany) TableName() string { return "companies" }

// IAMRolePermission links roles to permissions.
type IAMRolePermission struct {
	RoleID       uint `gorm:"primaryKey;not null;index"`
	PermissionID uint `gorm:"primaryKey;not null;index"`
}

func (IAMRolePermission) TableName() string { return "role_permissions" }
