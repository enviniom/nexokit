package queries

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company"

	"gorm.io/gorm"
)

func GetCompanyByPublicID(db *gorm.DB, publicID string) (*company.Company, error) {
	var c company.Company
	if err := db.Where("public_id = ?", publicID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
