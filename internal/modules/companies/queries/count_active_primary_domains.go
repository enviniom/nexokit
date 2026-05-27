package queries

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"gorm.io/gorm"
)

func CountActivePrimaryDomains(db *gorm.DB, companyID uint, excludeDomainPublicID string) (int64, error) {
	var n int64
	q := db.Model(&core.CompanyDomain{}).
		Where("company_id = ? AND kind = ? AND status = ?", companyID, core.CompanyDomainKindPrimary, core.CompanyDomainStatusActive)
	if excludeDomainPublicID != "" {
		q = q.Where("public_id <> ?", excludeDomainPublicID)
	}
	return n, q.Count(&n).Error
}
