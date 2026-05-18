package auth

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/users"
)

// RefreshToken stores a hashed opaque refresh token.
type RefreshToken struct {
	ID             uint   `gorm:"primaryKey"`
	PublicID       string `gorm:"type:char(26);uniqueIndex;not null"`
	UserID         uint   `gorm:"not null;index"`
	TokenHash      string `gorm:"size:64;not null;index"`
	ExpiresAt      time.Time
	RevokedAt      *time.Time
	ReplacedByHash *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	User           users.User
}
