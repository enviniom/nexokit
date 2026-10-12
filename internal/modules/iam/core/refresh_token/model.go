package refresh_token

import (
	"time"

	user "github.com/enviniom/nexokit/internal/modules/iam/core/auth"
)

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
	User           user.AuthUser
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
