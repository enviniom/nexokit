package create_role

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

type Repository interface {
	ExistsRoleByName(tc tenant.TenantContext, name string) (bool, error)
	ExistsRoleBySlug(tc tenant.TenantContext, slug string) (bool, error)
	Create(tc tenant.TenantContext, req core.CreateRoleRequest) (*core.IAMRole, error)
}

type gormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) ExistsRoleByName(tc tenant.TenantContext, name string) (bool, error) {
	return queries.ExistsRoleByName(r.db, tc, name, 0)
}

func (r *gormRepository) ExistsRoleBySlug(tc tenant.TenantContext, slug string) (bool, error) {
	return queries.ExistsRoleBySlug(r.db, tc, slug, 0)
}

func (r *gormRepository) Create(tc tenant.TenantContext, req core.CreateRoleRequest) (*core.IAMRole, error) {
	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}
	role := &core.IAMRole{BaseModel: shared.BaseModel{PublicID: publicID}, Name: req.Name, Slug: req.Slug, Description: req.Description, IsSystem: false}
	if !tc.IsRootScope {
		cid := tc.CompanyID
		role.CompanyID = &cid
	}
	if err := r.db.Create(role).Error; err != nil {
		return nil, err
	}
	return queries.GetRoleByPublicID(r.db.Preload("Company").Preload("Permissions"), tc, role.PublicID)
}
