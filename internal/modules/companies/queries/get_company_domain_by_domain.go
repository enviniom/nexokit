package queries

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"gorm.io/gorm"
)

func GetCompanyDomainByDomain(db *gorm.DB, domain string) (*core.CompanyDomain, error) {
	var d core.CompanyDomain
	if err := db.Where("domain = ?", domain).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}
