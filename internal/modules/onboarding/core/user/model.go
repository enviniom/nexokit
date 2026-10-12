package user

import "github.com/enviniom/nexokit/internal/shared"

type OnboardingUser struct {
	shared.BaseModel
	Name         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	RoleID       uint   `gorm:"not null"`
	CompanyID    *uint
	IsActive     bool `gorm:"not null;default:true"`
}

func (OnboardingUser) TableName() string {
	return "users"
}
