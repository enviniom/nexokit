package company

import "github.com/enviniom/nexokit/internal/shared"

type OnboardingCompany struct {
	shared.BaseModel
	Name   string `gorm:"not null"`
	Slug   string `gorm:"type:varchar(120);uniqueIndex;not null"`
	Status string `gorm:"type:varchar(20);not null;default:'active';index"`
}

func (OnboardingCompany) TableName() string {
	return "companies"
}
