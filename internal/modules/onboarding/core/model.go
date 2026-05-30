package core

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
