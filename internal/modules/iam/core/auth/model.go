package user

import "github.com/enviniom/nexokit/internal/shared"

// AuthRole is the local partial role model required by auth flows.
type AuthRole struct {
	shared.BaseModel
	Name string `gorm:"not null"`
	Slug string `gorm:"not null"`
}

func (AuthRole) TableName() string { return "roles" }

// AuthUser is the local partial user model required by auth flows.
type AuthUser struct {
	shared.BaseModel
	Name         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	RoleID       uint   `gorm:"not null"`
	CompanyID    *uint
	IsActive     bool `gorm:"not null;default:true"`
	Role         AuthRole
}

func (AuthUser) TableName() string { return "users" }

// SessionView is the sanitized projection used by the session read path.
// It intentionally has no persistence audit fields: /me historically emits zero timestamps.
type SessionView struct {
	PublicID    string
	Name        string
	Email       string
	IsActive    bool
	RoleID      uint
	RoleName    string
	RoleSlug    string
	CompanyID   *uint
	Permissions []string
}
