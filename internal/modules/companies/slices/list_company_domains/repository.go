package list_company_domains

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"

	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*company.Company, error)
	ListDomains(uint) ([]company_domain.CompanyDomain, error)
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
func (r *GormRepository) ListDomains(companyID uint) ([]company_domain.CompanyDomain, error) {
	var d []company_domain.CompanyDomain
	if err := r.db.Where("company_id = ?", companyID).Order("kind ASC, domain ASC").Find(&d).Error; err != nil {
		return nil, queries.MapCompanyDomainError(err)
	}
	return d, nil
}
