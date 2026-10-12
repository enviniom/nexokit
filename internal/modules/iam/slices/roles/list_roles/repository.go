package list_roles

import (
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]iamrole.RoleResponse, error)
	Count(tc tenant.TenantContext) (int64, error)
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) List(tc tenant.TenantContext, params query.ListParams) ([]iamrole.RoleResponse, error) {
	var roles []iamrole.IAMRole
	db := tenant.ApplyTenantScope(r.db, tc)
	db = gormutil.ApplySearch(db, params.Search, "name", "slug", "description")
	db = gormutil.ApplySorting(db, params.Sort, "created_at", "name", "slug")
	db = gormutil.ApplyPagination(db, params.Pagination.Page, params.Pagination.PerPage)
	if err := db.Preload("Company").Preload("Permissions").Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, err
	}

	res := make([]iamrole.RoleResponse, len(roles))
	for i := range roles {
		res[i] = toRoleResponse(&roles[i])
	}
	return res, nil
}

func (r *GormRepository) Count(tc tenant.TenantContext) (int64, error) {
	var count int64
	if err := tenant.ApplyTenantScope(r.db.Model(&iamrole.IAMRole{}), tc).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
