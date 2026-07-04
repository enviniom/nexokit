package rotate_token

import (
	"errors"
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/modules/auth/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByHash(hash string) (*core.RefreshToken, error)
	CreateRefreshToken(refresh *core.RefreshToken) error
	Revoke(hash string, replacedByHash *string) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) GetByHash(hash string) (*core.RefreshToken, error) {
	refresh, err := queries.FindRefreshTokenByHashWithUser(r.db, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrInvalidRefreshToken
		}
		return nil, err
	}
	return refresh, nil
}

func (r *GormRepository) CreateRefreshToken(refresh *core.RefreshToken) error {
	return r.db.Create(refresh).Error
}

func (r *GormRepository) Revoke(hash string, replacedByHash *string) error {
	updates := map[string]any{"revoked_at": time.Now()}
	if replacedByHash != nil {
		updates["replaced_by_hash"] = *replacedByHash
	}
	return r.db.Model(&core.RefreshToken{}).Where("token_hash = ?", hash).Updates(updates).Error
}
