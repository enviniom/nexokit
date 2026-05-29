package authenticate_user

import (
	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/modules/auth/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByEmail(email string) (*core.AuthUser, error)
	CreateRefreshToken(refresh *core.RefreshToken) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) GetByEmail(email string) (*core.AuthUser, error) {
	user, err := queries.FindUserByEmail(r.db, email)
	if err != nil {
		return nil, err
	}
	if err := r.db.First(&user.Role, user.RoleID).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *GormRepository) CreateRefreshToken(refresh *core.RefreshToken) error {
	return r.db.Create(refresh).Error
}
