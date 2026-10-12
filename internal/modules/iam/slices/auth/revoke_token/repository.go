package revoke_token

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/refresh_token"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"gorm.io/gorm"
	"time"
)

type Repository interface {
	GetByHash(hash string) (*refresh_token.RefreshToken, error)
	Revoke(hash string) error
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

func (r *GormRepository) Revoke(hash string) error {
	result := r.db.Model(&refresh_token.RefreshToken{}).Where("token_hash = ?", hash).Updates(map[string]any{"revoked_at": time.Now()})
	if result.Error != nil {
		return queries.MapRefreshTokenError(result.Error)
	}
	if result.RowsAffected == 0 {
		return queries.MapRefreshTokenError(gorm.ErrRecordNotFound)
	}
	return nil
}
