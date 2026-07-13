package update_company_domain

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*core.Company, error)
	GetDomainByPublicID(string) (*core.CompanyDomain, error)
	GetDomainByDomain(string) (*core.CompanyDomain, error)
	CountActivePrimaryDomains(uint, string) (int64, error)
	UpdateDomain(*core.CompanyDomain) error
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
func (r *GormRepository) GetDomainByPublicID(id string) (*core.CompanyDomain, error) {
	var d core.CompanyDomain
	if err := r.db.Where("public_id = ?", id).First(&d).Error; err != nil {
		return nil, queries.MapCompanyDomainError(err)
	}
	return &d, nil
}
func (r *GormRepository) GetDomainByDomain(name string) (*core.CompanyDomain, error) {
	domain, err := queries.GetCompanyDomainByDomain(r.db, name)
	if err != nil {
		return nil, queries.MapCompanyDomainError(err)
	}
	return domain, nil
}
func (r *GormRepository) CountActivePrimaryDomains(companyID uint, exclude string) (int64, error) {
	count, err := queries.CountActivePrimaryDomains(r.db, companyID, exclude)
	return count, queries.MapCompanyDomainError(err)
}
func (r *GormRepository) UpdateDomain(d *core.CompanyDomain) error {
	result := r.db.Model(&core.CompanyDomain{}).Where("public_id = ?", d.PublicID).Updates(map[string]any{"domain": d.Domain, "status": d.Status, "kind": d.Kind, "redirect_to_primary": d.RedirectToPrimary})
	if result.Error != nil {
		return queries.MapCompanyDomainError(result.Error)
	}
	if result.RowsAffected == 0 {
		return queries.MapCompanyDomainError(gorm.ErrRecordNotFound)
	}
	return nil
}
