package permission

import "github.com/enviniom/nexokit/internal/shared"

type OnboardingPermission struct {
	shared.BaseModel
	Slug         string `gorm:"type:varchar(120);uniqueIndex;not null"`
	Name         string `gorm:"type:varchar(120);not null"`
	Module       string `gorm:"type:varchar(80);not null;index"`
	Action       string `gorm:"type:varchar(80);not null"`
	Description  string
	IsSystem     bool `gorm:"not null;default:false"`
	DisplayOrder int  `gorm:"not null;default:0;index"`
}

func (OnboardingPermission) TableName() string {
	return "permissions"
}
