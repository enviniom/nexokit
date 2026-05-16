package users

import (
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/shared"
)

// User represents a user in the system.
type User struct {
	shared.BaseModel
	Name         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	RoleID       uint   `gorm:"not null"`
	CompanyID    *uint
	IsActive     bool `gorm:"not null;default:true"`
	Role         roles.Role
}
