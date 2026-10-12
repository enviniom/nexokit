package role

import "github.com/enviniom/nexokit/internal/shared"

type OnboardingRole struct {
	shared.BaseModel
	Name        string `gorm:"not null"`
	Slug        string `gorm:"not null"`
	CompanyID   *uint
	Description string
	IsSystem    bool `gorm:"not null;default:false"`
}

func (OnboardingRole) TableName() string {
	return "roles"
}
