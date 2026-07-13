package create_company_domain

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*core.Company, error)
	GetDomainByDomain(string) (*core.CompanyDomain, error)
	CountActivePrimaryDomains(uint, string) (int64, error)
	CreateDomain(*core.CompanyDomain) error
}
type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
func (r *GormRepository) GetByPublicID(id string) (*core.Company, error) {
	c, err := queries.GetCompanyByPublicID(r.db, id)
	if err != nil {
		return nil, queries.MapCompanyError(err)
	}
	return c, nil
}
func (r *GormRepository) GetDomainByDomain(d string) (*core.CompanyDomain, error) {
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
func (r *GormRepository) CreateDomain(d *core.CompanyDomain) error {
	return queries.MapCompanyDomainError(r.db.Create(d).Error)
}
