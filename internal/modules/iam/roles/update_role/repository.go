package update_role

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository interface {
	GetRoleByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMRole, error)
	ExistsRoleByName(tc tenant.TenantContext, name string, excludeRoleID uint) (bool, error)
	ExistsRoleBySlug(tc tenant.TenantContext, slug string, excludeRoleID uint) (bool, error)
	UpdateRole(tc tenant.TenantContext, role *core.IAMRole) error
}

type gormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) GetRoleByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMRole, error) {
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

func (r *gormRepository) ExistsRoleByName(tc tenant.TenantContext, name string, excludeRoleID uint) (bool, error) {
	return queries.ExistsRoleByName(r.db, tc, name, excludeRoleID)
}

func (r *gormRepository) ExistsRoleBySlug(tc tenant.TenantContext, slug string, excludeRoleID uint) (bool, error) {
	return queries.ExistsRoleBySlug(r.db, tc, slug, excludeRoleID)
}

func (r *gormRepository) UpdateRole(tc tenant.TenantContext, role *core.IAMRole) error {
	res := tenant.ApplyTenantScope(r.db.Model(&core.IAMRole{}), tc).
		Where("id = ?", role.ID).
		Updates(map[string]any{"name": role.Name, "slug": role.Slug, "description": role.Description})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return core.ErrNotFound
	}
	return nil
}
