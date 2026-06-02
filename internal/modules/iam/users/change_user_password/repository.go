package change_user_password

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns persistence for the change_user_password slice.
type Repository interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMUser, error)
	GetRoleBySlug(slug string) (*core.IAMRole, error)
	UpdatePassword(userID uint, hash string) error
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new change_user_password repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMUser, error) {
	user, err := queries.GetUserByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *GormRepository) GetRoleBySlug(slug string) (*core.IAMRole, error) {
	var role core.IAMRole
	if err := r.db.Where("slug = ?", slug).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *GormRepository) UpdatePassword(userID uint, hash string) error {
	return r.db.Model(&core.IAMUser{}).Where("id = ?", userID).Update("password_hash", hash).Error
}
