package auth

import (
	"time"

	"gorm.io/gorm"
)

// RefreshRepository defines persistence for refresh tokens.
type RefreshRepository interface {
	Create(refresh *RefreshToken) error
	GetByHash(hash string) (*RefreshToken, error)
	Revoke(hash string, replacedByHash *string) error
}

// GormRefreshRepository is the GORM implementation of RefreshRepository.
type GormRefreshRepository struct {
	db *gorm.DB
}

// NewRefreshRepository creates a refresh token repository.
func NewRefreshRepository(db *gorm.DB) RefreshRepository {
	return &GormRefreshRepository{db: db}
}

// Create stores a refresh token hash.
func (r *GormRefreshRepository) Create(refresh *RefreshToken) error {
	return r.db.Create(refresh).Error
}

// GetByHash returns a refresh token by hash with user and role preloaded.
func (r *GormRefreshRepository) GetByHash(hash string) (*RefreshToken, error) {
	var refresh RefreshToken
	if err := r.db.Preload("User.Role").Where("token_hash = ?", hash).First(&refresh).Error; err != nil {
		return nil, err
	}
	return &refresh, nil
}

// Revoke marks a refresh token as revoked and records the replacement hash when rotating.
func (r *GormRefreshRepository) Revoke(hash string, replacedByHash *string) error {
	updates := map[string]any{"revoked_at": time.Now()}
	if replacedByHash != nil {
		updates["replaced_by_hash"] = *replacedByHash
	}
	return r.db.Model(&RefreshToken{}).Where("token_hash = ?", hash).Updates(updates).Error
}
