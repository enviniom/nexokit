package view_company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"

	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*company.Company, error)
}
type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*company.Company, error) {
	var c company.Company
	query := r.db
	if !tc.IsRootScope {
		query = query.Where("id = ?", tc.CompanyID)
	}
	if err := query.Preload("Domains", func(db *gorm.DB) *gorm.DB { return db.Order("kind ASC, domain ASC") }).Where("public_id = ?", publicID).First(&c).Error; err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return &c, nil
}
