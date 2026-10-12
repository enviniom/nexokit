package view_role_permission_catalog

import (
	"errors"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository interface {
	GetRoleByPublicID(tc tenant.TenantContext, publicID string) (*iamrole.IAMRole, error)
	ListPermissionCatalog() ([]permission.IAMPermission, error)
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetRoleByPublicID(tc tenant.TenantContext, publicID string) (*iamrole.IAMRole, error) {
	var role iamrole.IAMRole
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

func (r *GormRepository) ListPermissionCatalog() ([]permission.IAMPermission, error) {
	var items []permission.IAMPermission
	if err := r.db.Order("module ASC, display_order ASC, slug ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
