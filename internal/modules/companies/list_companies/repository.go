package list_companies

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
	"gorm.io/gorm"
)

type Repository interface {
	List(req core.ListCompaniesRequest) ([]core.Company, int64, error)
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) List(req core.ListCompaniesRequest) ([]core.Company, int64, error) {
	var rows []core.Company
	var total int64
	db := r.db.Model(&core.Company{})
	if !req.IncludeInactive {
		db = db.Where("status = ?", core.CompanyStatusActive)
	}
	db = gormutil.ApplyStatusFilter(db, req.ListParams.Filters, "status")
	db = gormutil.ApplyDateRange(db, req.ListParams.Filters, "created_at")
	db = gormutil.ApplySearch(db, req.ListParams.Search, "name", "slug")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	sort := req.ListParams.Sort
	if sort.Sort == "" {
		sort.Sort = "created_at"
	}
	if sort.Order == "" {
		sort.Order = "desc"
	}
	db = gormutil.ApplySorting(db, sort, "created_at", "name", "slug", "status")
	db = gormutil.ApplyPagination(db, req.ListParams.Pagination.Page, req.ListParams.Pagination.PerPage)
	if err := db.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
