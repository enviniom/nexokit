package users

import (
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository defines the persistence contract for users.
type Repository interface {
	List(tc tenant.TenantContext, params query.ListParams) ([]User, error)
	Count(tc tenant.TenantContext, params query.ListParams) (int64, error)
	GetByPublicID(tc tenant.TenantContext, publicID string) (*User, error)
	GetAuthUser(publicID string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(tc tenant.TenantContext, publicID string) error
	ListPublicIDsByRoleID(roleID uint) ([]string, error)
}

// GormRepository is the GORM implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM users repository.
func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// List returns paginated users.
func (r *GormRepository) List(tc tenant.TenantContext, params query.ListParams) ([]User, error) {
	var users []User
	db := applyUserListFilters(tenant.ApplyTenantScope(r.db, tc), params)
	db = gormutil.ApplySorting(db, withDefaultUserSort(params.Sort), "created_at", "name", "email")
	db = gormutil.ApplyPagination(db, params.Pagination.Page, params.Pagination.PerPage)
	if err := db.Preload("Role").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Count returns the total number of users.
func (r *GormRepository) Count(tc tenant.TenantContext, params query.ListParams) (int64, error) {
	var count int64
	db := applyUserListFilters(tenant.ApplyTenantScope(r.db.Model(&User{}), tc), params)
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func applyUserListFilters(db *gorm.DB, params query.ListParams) *gorm.DB {
	db = applyUserStatusFilter(db, params.Filters.Status)
	db = gormutil.ApplyDateRange(db, params.Filters, "created_at")
	return gormutil.ApplySearch(db, params.Search, "name", "email")
}

func applyUserStatusFilter(db *gorm.DB, status string) *gorm.DB {
	switch status {
	case "active":
		return db.Where("is_active = ?", true)
	case "inactive":
		return db.Where("is_active = ?", false)
	default:
		return db
	}
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

// GetByPublicID returns a user by its public ID.
func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*User, error) {
	var user User
	db := tenant.ApplyTenantScope(r.db.Preload("Role"), tc)
	if err := db.Where("public_id = ?", publicID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// GetAuthUser returns a user by public ID without tenant scope for auth bootstrap.
func (r *GormRepository) GetAuthUser(publicID string) (*User, error) {
	var user User
	if err := r.db.Preload("Role").Where("public_id = ?", publicID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail returns a user by its email.
func (r *GormRepository) GetByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// Create persists a new user.
func (r *GormRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// Update modifies an existing user.
func (r *GormRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// Delete soft-deletes a user by its public ID.
func (r *GormRepository) Delete(tc tenant.TenantContext, publicID string) error {
	db := tenant.ApplyTenantScope(r.db, tc)
	result := db.Where("public_id = ?", publicID).Delete(&User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListPublicIDsByRoleID returns public IDs for users assigned to the role.
func (r *GormRepository) ListPublicIDsByRoleID(roleID uint) ([]string, error) {
	var publicIDs []string
	err := r.db.Model(&User{}).
		Where("role_id = ?", roleID).
		Order("public_id ASC").
		Pluck("public_id", &publicIDs).Error
	return publicIDs, err
}
