package rotate_token

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core/refresh_token"
	"github.com/enviniom/nexokit/internal/modules/auth/queries"
	"gorm.io/gorm"
	"time"
)

type Repository interface {
	GetByHash(hash string) (*refresh_token.RefreshToken, error)
	CreateRefreshToken(refresh *refresh_token.RefreshToken) error
	Revoke(hash string, replacedByHash *string) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) GetByHash(hash string) (*refresh_token.RefreshToken, error) {
	refresh, err := queries.FindRefreshTokenByHashWithUser(r.db, hash)
	if err != nil {
		return nil, queries.MapRefreshTokenError(err)
	}
	return refresh, nil
}

func (r *GormRepository) CreateRefreshToken(refresh *refresh_token.RefreshToken) error {
	return queries.MapRefreshTokenError(r.db.Create(refresh).Error)
}

func (r *GormRepository) Revoke(hash string, replacedByHash *string) error {
	updates := map[string]any{"revoked_at": time.Now()}
	if replacedByHash != nil {
		updates["replaced_by_hash"] = *replacedByHash
	}
	result := r.db.Model(&refresh_token.RefreshToken{}).Where("token_hash = ?", hash).Updates(updates)
	if result.Error != nil {
		return queries.MapRefreshTokenError(result.Error)
	}
	if result.RowsAffected == 0 {
		return queries.MapRefreshTokenError(gorm.ErrRecordNotFound)
	}
	return nil
}
