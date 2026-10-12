package create_company_domain

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"

	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*company.Company, error)
	GetDomainByDomain(string) (*company_domain.CompanyDomain, error)
	CountActivePrimaryDomains(uint, string) (int64, error)
	CreateDomain(*company_domain.CompanyDomain) error
}
type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
func (r *GormRepository) GetByPublicID(id string) (*company.Company, error) {
	c, err := queries.GetCompanyByPublicID(r.db, id)
	if err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return c, nil
}
func (r *GormRepository) GetDomainByDomain(d string) (*company_domain.CompanyDomain, error) {
	domain, err := queries.GetCompanyDomainByDomain(r.db, d)
	if err != nil {
		return nil, queries.MapCompanyDomainError(err)
	}
	return domain, nil
}
func (r *GormRepository) CountActivePrimaryDomains(companyID uint, _ string) (int64, error) {
	count, err := queries.CountActivePrimaryDomains(r.db, companyID, "")
	return count, queries.MapCompanyDomainError(err)
}
func (r *GormRepository) CreateDomain(d *company_domain.CompanyDomain) error {
	return queries.MapCompanyDomainError(r.db.Create(d).Error)
}
