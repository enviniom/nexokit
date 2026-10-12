package user

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/shared"
)

// IAMUser is the IAM-local user model used to avoid cross-module imports.
type IAMUser struct {
	shared.BaseModel
	Name         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	RoleID       uint   `gorm:"not null"`
	CompanyID    *uint
	IsActive     bool            `gorm:"not null;default:true"`
	Role         iamrole.IAMRole `gorm:"foreignKey:RoleID"`
}

func (IAMUser) TableName() string { return "users" }
