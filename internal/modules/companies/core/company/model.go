package company

import (
	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	internalshared "github.com/enviniom/nexokit/internal/shared"
)

type Company struct {
	internalshared.BaseModel
	Name    string                         `gorm:"not null"`
	Slug    string                         `gorm:"type:varchar(120);uniqueIndex;not null"`
	Status  string                         `gorm:"type:varchar(20);not null;default:'active';index"`
	Domains []company_domain.CompanyDomain `gorm:"foreignKey:CompanyID"`
}

func (Company) TableName() string { return "companies" }
