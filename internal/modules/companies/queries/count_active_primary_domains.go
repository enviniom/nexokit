package queries

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"gorm.io/gorm"
)

func CountActivePrimaryDomains(db *gorm.DB, companyID uint, excludeDomainPublicID string) (int64, error) {
	var n int64
	q := db.Model(&company_domain.CompanyDomain{}).
		Where("company_id = ? AND kind = ? AND status = ?", companyID, core.CompanyDomainKindPrimary, core.CompanyDomainStatusActive)
	if excludeDomainPublicID != "" {
		q = q.Where("public_id <> ?", excludeDomainPublicID)
	}
	return n, q.Count(&n).Error
}
