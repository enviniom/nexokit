package core

import (
	"time"

	"github.com/enviniom/nexokit/internal/shared"
)

// AuthRole is the local partial role model required by auth flows.
type AuthRole struct {
	shared.BaseModel
	Name string `gorm:"not null"`
	Slug string `gorm:"not null"`
}

// TableName maps AuthRole to the existing roles table.
func (AuthRole) TableName() string {
	return "roles"
}

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

// TableName maps AuthUser to the existing users table.
func (AuthUser) TableName() string {
	return "users"
}

// RefreshToken stores a hashed opaque refresh token.
type RefreshToken struct {
	ID             uint `gorm:"primaryKey"`
	PublicID       string
	UserID         uint `gorm:"not null;index"`
	TokenHash      string
	ExpiresAt      time.Time
	RevokedAt      *time.Time
	ReplacedByHash *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	User           AuthUser
}

// TableName maps RefreshToken to the existing refresh_tokens table.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
