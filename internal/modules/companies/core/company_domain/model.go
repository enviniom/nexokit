package company_domain

import internalshared "github.com/enviniom/nexokit/internal/shared"

// CompanyAssociation is the local companies-table record used for domain preloading.
// It avoids a bidirectional dependency on the company entity.
type CompanyAssociation struct {
	internalshared.BaseModel
	Name   string `gorm:"not null"`
	Slug   string `gorm:"type:varchar(120);uniqueIndex;not null"`
	Status string `gorm:"type:varchar(20);not null;default:'active';index"`
}

func (CompanyAssociation) TableName() string { return "companies" }

type CompanyDomain struct {
	internalshared.BaseModel
	CompanyID         uint               `gorm:"not null;index"`
	Company           CompanyAssociation `gorm:"foreignKey:CompanyID"`
	Domain            string             `gorm:"type:varchar(255);uniqueIndex;not null"`
	Status            string             `gorm:"type:varchar(40);not null;default:'active';index"`
	Kind              string             `gorm:"type:varchar(40);not null;index"`
	RedirectToPrimary bool               `gorm:"not null;default:false"`
}

func (CompanyDomain) TableName() string { return "company_domains" }
