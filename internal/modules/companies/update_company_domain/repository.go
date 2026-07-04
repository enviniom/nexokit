package update_company_domain

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"github.com/enviniom/nexokit/internal/platform/gormutil"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrCompanyNotFound
		}
		return nil, err
	}
	return c, nil
}
func (r *GormRepository) GetDomainByPublicID(id string) (*core.CompanyDomain, error) {
	var d core.CompanyDomain
	if err := r.db.Where("public_id = ?", id).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrCompanyDomainNotFound
		}
		return nil, err
	}
	return &d, nil
}
func (r *GormRepository) GetDomainByDomain(name string) (*core.CompanyDomain, error) {
	domain, err := queries.GetCompanyDomainByDomain(r.db, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrCompanyDomainNotFound
		}
		return nil, err
	}
	return domain, nil
}
func (r *GormRepository) CountActivePrimaryDomains(companyID uint, exclude string) (int64, error) {
	return queries.CountActivePrimaryDomains(r.db, companyID, exclude)
}
func (r *GormRepository) UpdateDomain(d *core.CompanyDomain) error {
	if err := r.db.Save(d).Error; err != nil {
		if gormutil.IsUniqueConstraintError(err) {
			return core.ErrDuplicateCompanyDomain
		}
		return err
	}
	return nil
}
