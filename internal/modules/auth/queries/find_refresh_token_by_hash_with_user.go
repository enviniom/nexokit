package queries

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"gorm.io/gorm"
)

// FindRefreshTokenByHashWithUser returns one refresh token by hash with user and role preloaded.
func FindRefreshTokenByHashWithUser(db *gorm.DB, hash string) (*core.RefreshToken, error) {
	var refresh core.RefreshToken
	if err := db.Preload("User.Role").Where("token_hash = ?", hash).First(&refresh).Error; err != nil {
		return nil, err
	}
	return &refresh, nil
}
