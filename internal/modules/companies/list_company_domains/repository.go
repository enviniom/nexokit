package list_company_domains

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/queries"
	"gorm.io/gorm"
)

type Repository interface {
	GetByPublicID(string) (*core.Company, error)
	ListDomains(uint) ([]core.CompanyDomain, error)
}
type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
func (r *GormRepository) GetByPublicID(id string) (*core.Company, error) {
	return queries.GetCompanyByPublicID(r.db, id)
}
func (r *GormRepository) ListDomains(companyID uint) ([]core.CompanyDomain, error) {
	var d []core.CompanyDomain
	if err := r.db.Where("company_id = ?", companyID).Order("kind ASC, domain ASC").Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}
