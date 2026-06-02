package core

import "github.com/enviniom/nexokit/internal/shared"

// IAMUser is the IAM-local user model used to avoid cross-module imports.
type IAMUser struct {
	shared.BaseModel
	Name         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	RoleID       uint   `gorm:"not null"`
	CompanyID    *uint
	IsActive     bool    `gorm:"not null;default:true"`
	Role         IAMRole `gorm:"foreignKey:RoleID"`
}

func (IAMUser) TableName() string { return "users" }

// IAMRole is the IAM-local role model used to avoid cross-module imports.
type IAMRole struct {
	shared.BaseModel
	Name        string `gorm:"not null"`
	Slug        string `gorm:"not null"`
	CompanyID   *uint
	Company     IAMCompany `gorm:"foreignKey:CompanyID"`
	Description string
	IsSystem    bool            `gorm:"not null;default:false"`
	Permissions []IAMPermission `gorm:"many2many:role_permissions;joinForeignKey:RoleID;joinReferences:PermissionID"`
}

func (IAMRole) TableName() string { return "roles" }

// IAMPermission is the IAM-local permission model.
type IAMPermission struct {
	shared.BaseModel
	Slug         string `gorm:"type:varchar(120);uniqueIndex;not null"`
	Name         string `gorm:"type:varchar(120);not null"`
	Module       string `gorm:"type:varchar(80);not null;index"`
	Action       string `gorm:"type:varchar(80);not null"`
	Description  string
	IsSystem     bool `gorm:"not null;default:false"`
	DisplayOrder int  `gorm:"not null;default:0;index"`
}

func (IAMPermission) TableName() string { return "permissions" }

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

func (IAMRolePermission) TableName() string {
	return "role_permissions"
}
