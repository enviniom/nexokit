package queries

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"

	"gorm.io/gorm"
)

func GetCompanyDomainByDomain(db *gorm.DB, domain string) (*company_domain.CompanyDomain, error) {
	var d company_domain.CompanyDomain
	if err := db.Where("domain = ?", domain).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}
