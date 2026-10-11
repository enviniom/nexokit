package authenticate_user

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core/refresh_token"
	"github.com/enviniom/nexokit/internal/modules/auth/core/user"
	"github.com/enviniom/nexokit/internal/modules/auth/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByEmail(email string) (*user.AuthUser, error)
	CreateRefreshToken(refresh *refresh_token.RefreshToken) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) GetByEmail(email string) (*user.AuthUser, error) {
	user, err := queries.FindUserByEmail(r.db, email)
	if err != nil {
		return nil, queries.MapUserError(err)
	}
	if err := r.db.First(&user.Role, user.RoleID).Error; err != nil {
		return nil, queries.MapUserError(err)
	}
	return user, nil
}

func (r *GormRepository) CreateRefreshToken(refresh *refresh_token.RefreshToken) error {
	return queries.MapRefreshTokenError(r.db.Create(refresh).Error)
}
