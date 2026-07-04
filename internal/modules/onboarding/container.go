package onboarding

import (
	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/modules/onboarding/onboard_company"
	"gorm.io/gorm"
)

type Container struct {
	OnboardCompany *onboard_company.Handler
}

type Config struct {
	PasswordHasher core.PasswordHasher
	PlatformDomain string
}

func NewContainer(db *gorm.DB, cfg Config) *Container {
	repo := onboard_company.NewRepository(db)
	service := onboard_company.NewService(repo, cfg.PasswordHasher, cfg.PlatformDomain)
	handler := onboard_company.NewHandler(service)

	return &Container{OnboardCompany: handler}
}
