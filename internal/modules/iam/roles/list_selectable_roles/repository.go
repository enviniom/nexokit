package list_selectable_roles

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository interface {
	List(tc tenant.TenantContext) ([]response.SelectResponse, error)
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) List(tc tenant.TenantContext) ([]response.SelectResponse, error) {
	var roles []core.IAMRole
	if err := tenant.ApplyTenantScope(r.db, tc).
		Preload("Company").
		Where("slug <> ?", core.RootRoleSlug).
		Order("name ASC").
		Find(&roles).Error; err != nil {
		return nil, err
	}

	items := make([]response.SelectResponse, len(roles))
	for i, role := range roles {
		meta := map[string]any{"slug": role.Slug}
		if role.CompanyID != nil && role.Company.PublicID != "" {
			meta["company_id"] = role.Company.PublicID
		}
		items[i] = response.SelectResponse{ID: role.PublicID, Name: role.Name, Meta: meta}
	}

	return items, nil
}
