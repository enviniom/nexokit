package view_role_permission_catalog

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository interface {
	GetRoleByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMRole, error)
	ListPermissionCatalog() ([]core.IAMPermission, error)
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetRoleByPublicID(tc tenant.TenantContext, publicID string) (*core.IAMRole, error) {
	var role core.IAMRole
	err := tenant.ApplyTenantScope(r.db.Preload("Permissions"), tc).
		Where("public_id = ?", publicID).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *GormRepository) ListPermissionCatalog() ([]core.IAMPermission, error) {
	var items []core.IAMPermission
	if err := r.db.Order("module ASC, display_order ASC, slug ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
