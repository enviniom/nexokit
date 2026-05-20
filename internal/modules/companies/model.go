package companies

import "github.com/enviniom/nexokit/internal/shared"

const (
	CompanyStatusActive   = "active"
	CompanyStatusInactive = "inactive"
)

// Company represents an organization tenant in the system.
type Company struct {
	shared.BaseModel
	Name      string  `gorm:"not null"`
	Slug      string  `gorm:"type:varchar(120);uniqueIndex;not null"`
	Domain    *string `gorm:"type:varchar(255);index"`
	Subdomain *string `gorm:"type:varchar(120);index"`
	Status    string  `gorm:"type:varchar(20);not null;default:'active';index"`
}
