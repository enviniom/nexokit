package list_users

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository owns the list/count persistence for the list_users slice.
type Repository interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]core.UserResponse, error)
	Count(tc tenant.TenantContext, params query.ListParams) (int64, error)
}

// GormRepository is the GORM-backed implementation of Repository.
type GormRepository struct{ db *gorm.DB }

// NewRepository creates a new list_users repository.
func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) List(tc tenant.TenantContext, params query.ListParams) ([]core.UserResponse, error) {
	var users []core.IAMUser
	db := applyUserListFilters(tenant.ApplyTenantScope(r.db, tc), params)
	db = gormutil.ApplySorting(db, withDefaultUserSort(params.Sort), "created_at", "name", "email")
	db = gormutil.ApplyPagination(db, params.Pagination.Page, params.Pagination.PerPage)
	if err := db.Preload("Role").Find(&users).Error; err != nil {
		return nil, err
	}
	res := make([]core.UserResponse, len(users))
	for i := range users {
		res[i] = toUserResponse(&users[i])
	}
	return res, nil
}

func (r *GormRepository) Count(tc tenant.TenantContext, params query.ListParams) (int64, error) {
	var count int64
	db := applyUserListFilters(tenant.ApplyTenantScope(r.db.Model(&core.IAMUser{}), tc), params)
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func toUserResponse(u *core.IAMUser) core.UserResponse {
	return core.UserResponse{
		PublicID:  u.PublicID,
		Name:      u.Name,
		Email:     u.Email,
		IsActive:  u.IsActive,
		RoleID:    u.RoleID,
		RoleName:  u.Role.Name,
		CompanyID: u.CompanyID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		CreatedBy: u.CreatedBy,
		UpdatedBy: u.UpdatedBy,
	}
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
