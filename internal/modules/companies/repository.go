package companies

import (
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

// Repository defines the persistence contract for companies.
type Repository interface {
	List(req ListCompaniesRequest) ([]Company, int64, error)
	GetByPublicID(publicID string) (*Company, error)
	GetBySlugIncludingDeleted(slug string) (*Company, error)
	Create(company *Company) error
	Update(company *Company) error
	Delete(publicID string) error
}

// GormRepository is the GORM implementation of Repository.
type GormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM companies repository.
func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) List(req ListCompaniesRequest) ([]Company, int64, error) {
	var companies []Company
	var total int64
	db := applyCompanyListFilters(r.db.Model(&Company{}), req)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	db = gormutil.ApplySorting(db, withDefaultCompanySort(req.ListParams.Sort), "created_at", "name", "slug", "status")
	db = gormutil.ApplyPagination(db, req.ListParams.Pagination.Page, req.ListParams.Pagination.PerPage)
	if err := db.Find(&companies).Error; err != nil {
		return nil, 0, err
	}
	return companies, total, nil
}

func applyCompanyListFilters(db *gorm.DB, req ListCompaniesRequest) *gorm.DB {
	if !req.IncludeInactive {
		db = db.Where("status = ?", CompanyStatusActive)
	}
	db = gormutil.ApplyStatusFilter(db, req.ListParams.Filters, "status")
	db = gormutil.ApplyDateRange(db, req.ListParams.Filters, "created_at")
	return gormutil.ApplySearch(db, req.ListParams.Search, "name", "slug")
}

func withDefaultCompanySort(sort query.SortParams) query.SortParams {
	if sort.Sort == "" {
		sort.Sort = "created_at"
	}
	if sort.Order == "" {
		sort.Order = "desc"
	}
	return sort
}

func (r *GormRepository) GetByPublicID(publicID string) (*Company, error) {
	var company Company
	if err := r.db.Where("public_id = ?", publicID).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *GormRepository) GetBySlugIncludingDeleted(slug string) (*Company, error) {
	var company Company
	if err := r.db.Unscoped().Where("slug = ?", slug).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *GormRepository) Create(company *Company) error {
	return r.db.Create(company).Error
}

func (r *GormRepository) Update(company *Company) error {
	return r.db.Save(company).Error
}

func (r *GormRepository) Delete(publicID string) error {
	return r.db.Where("public_id = ?", publicID).Delete(&Company{}).Error
}

// FindByPublicIDOrSlug implements middleware.CompanyResolver.
func (r *GormRepository) FindByPublicIDOrSlug(value string) (tenant.CompanyRef, error) {
	var company Company
	if err := r.db.Where("public_id = ? OR slug = ?", value, value).First(&company).Error; err != nil {
		return tenant.CompanyRef{}, err
	}
	return tenant.CompanyRef{ID: company.ID, Slug: company.Slug}, nil
}

// FindByHost implements middleware.CompanyResolver.
func (r *GormRepository) FindByHost(host string) (tenant.CompanyRef, error) {
	var company Company
	if err := r.db.Where("domain = ?", host).First(&company).Error; err != nil {
		return tenant.CompanyRef{}, err
	}
	return tenant.CompanyRef{ID: company.ID, Slug: company.Slug}, nil
}
