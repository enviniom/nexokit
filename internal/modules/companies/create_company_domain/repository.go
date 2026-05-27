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
	return queries.GetCompanyByPublicID(r.db, id)
}
func (r *GormRepository) GetDomainByDomain(d string) (*core.CompanyDomain, error) {
	return queries.GetCompanyDomainByDomain(r.db, d)
}
func (r *GormRepository) CountActivePrimaryDomains(companyID uint, _ string) (int64, error) {
	return queries.CountActivePrimaryDomains(r.db, companyID, "")
}
func (r *GormRepository) CreateDomain(d *core.CompanyDomain) error {
	return r.db.Create(d).Error
}
