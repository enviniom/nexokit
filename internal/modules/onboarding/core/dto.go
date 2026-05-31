package core

import (
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// OnboardCompanyRequest carries the required payload to onboard a company and provision its administrator.
type OnboardCompanyRequest struct {
	Name                    string  `json:"name"`
	Slug                    string  `json:"slug"`
	Domain                  *string `json:"domain,omitempty"`
	GenerateTechnicalDomain bool    `json:"generate_technical_domain"`
	AdminName               string  `json:"admin_name"`
	AdminEmail              string  `json:"admin_email"`
	AdminPassword           string  `json:"admin_password"`
}

// OnboardCompanyResponse holds the output values after successful onboarding.
type OnboardCompanyResponse struct {
	CompanyPublicID string `json:"company_public_id"`
	CompanySlug     string `json:"company_slug"`
	AdminPublicID   string `json:"admin_public_id"`
	AdminEmail      string `json:"admin_email"`
}

// Validate executes structured, field-level validations on the onboarding request.
func (r OnboardCompanyRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "admin_name", r.AdminName).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "admin_email", r.AdminEmail).Required().Apply(validator.ValidEmail())
	validator.Field(errs, "admin_password", r.AdminPassword).Required().Apply(validator.MinLength(8))
	return errs
}
