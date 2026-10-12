package list_users

import (
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns the list/count persistence for the list_users slice.
type Repository interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]iamuser.UserResponse, error)
	Count(tc tenant.TenantContext, params query.ListParams) (int64, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new list_users repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) List(tc tenant.TenantContext, params query.ListParams) ([]iamuser.UserResponse, error) {
	var users []iamuser.IAMUser
	db := applyUserListFilters(tenant.ApplyTenantScope(r.db, tc), params)
	db = gormutil.ApplySorting(db, withDefaultUserSort(params.Sort), "created_at", "name", "email")
	db = gormutil.ApplyPagination(db, params.Pagination.Page, params.Pagination.PerPage)
	if err := db.Preload("Role").Find(&users).Error; err != nil {
		return nil, err
	}
	res := make([]iamuser.UserResponse, len(users))
	for i := range users {
		res[i].FromModel(&users[i])
	}
	return res, nil
}

func (r *GormRepository) Count(tc tenant.TenantContext, params query.ListParams) (int64, error) {
	var count int64
	db := applyUserListFilters(tenant.ApplyTenantScope(r.db.Model(&iamuser.IAMUser{}), tc), params)
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func applyUserListFilters(db *gorm.DB, params query.ListParams) *gorm.DB {
	switch params.Filters.Status {
	case "active":
		db = db.Where("is_active = ?", true)
	case "inactive":
		db = db.Where("is_active = ?", false)
	}
	db = gormutil.ApplyDateRange(db, params.Filters, "created_at")
	return gormutil.ApplySearch(db, params.Search, "name", "email")
}

func withDefaultUserSort(sort query.SortParams) query.SortParams {
	if sort.Sort == "" {
		sort.Sort = "created_at"
	}
	if sort.Order == "" {
		sort.Order = "desc"
	}
	return sort
}
