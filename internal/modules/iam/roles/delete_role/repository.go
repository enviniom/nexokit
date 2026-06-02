package delete_role

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMRole, error)
	CountUsersByRoleID(roleID uint) (int64, error)
	DeleteByPublicID(tc tenant.TenantContext, publicID string) error
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMRole, error) {
	// NOTE: this wrapper delegates query details to queries.GetRoleByPublicID.
	// Full query behavior coverage belongs to internal/modules/iam/queries tests.
	role, err := queries.GetRoleByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return role, nil
}

func (r *GormRepository) CountUsersByRoleID(roleID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&core.IAMUser{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormRepository) DeleteByPublicID(tc tenant.TenantContext, publicID string) error {
	res := tenant.ApplyTenantScope(r.db, tc).Where("public_id = ?", publicID).Delete(&core.IAMRole{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return core.ErrNotFound
	}
	return nil
}
