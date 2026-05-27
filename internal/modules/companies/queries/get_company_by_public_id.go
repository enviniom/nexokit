package queries

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core"
	"gorm.io/gorm"
)

func GetCompanyByPublicID(db *gorm.DB, publicID string) (*core.Company, error) {
	var c core.Company
	if err := db.Where("public_id = ?", publicID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
