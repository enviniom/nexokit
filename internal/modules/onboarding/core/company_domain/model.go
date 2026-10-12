package company_domain

import "github.com/enviniom/nexokit/internal/shared"

type OnboardingCompanyDomain struct {
	shared.BaseModel
	CompanyID         uint   `gorm:"not null;index"`
	Domain            string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Status            string `gorm:"type:varchar(40);not null;default:'active';index"`
	Kind              string `gorm:"type:varchar(40);not null;index"`
	RedirectToPrimary bool   `gorm:"not null;default:false"`
}

func (OnboardingCompanyDomain) TableName() string {
	return "company_domains"
}
