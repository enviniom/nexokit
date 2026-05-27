package companies

import "github.com/enviniom/nexokit/internal/shared"

const (
	CompanyStatusActive   = "active"
	CompanyStatusInactive = "inactive"
)

const (
	CompanyDomainStatusActive              = "active"
	CompanyDomainStatusInactive            = "inactive"
	CompanyDomainStatusPendingVerification = "pending_verification"
)

const (
	CompanyDomainKindPrimary   = "primary"
	CompanyDomainKindAlias     = "alias"
	CompanyDomainKindTechnical = "technical"
)

// Company represents an organization tenant in the system.
type Company struct {
	shared.BaseModel
	Name    string          `gorm:"not null"`
	Slug    string          `gorm:"type:varchar(120);uniqueIndex;not null"`
	Status  string          `gorm:"type:varchar(20);not null;default:'active';index"`
	Domains []CompanyDomain `gorm:"foreignKey:CompanyID"`
}

// CompanyDomain represents a hostname owned by a company tenant.
type CompanyDomain struct {
	shared.BaseModel
	CompanyID         uint    `gorm:"not null;index"`
	Company           Company `gorm:"foreignKey:CompanyID"`
	Domain            string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Status            string  `gorm:"type:varchar(40);not null;default:'active';index"`
	Kind              string  `gorm:"type:varchar(40);not null;index"`
	RedirectToPrimary bool    `gorm:"not null;default:false"`
}
