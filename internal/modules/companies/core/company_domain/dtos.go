package company_domain

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/shared/string"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

type CompanyDomainResponse struct {
	PublicID          string    `json:"public_id"`
	CompanyPublicID   string    `json:"company_public_id"`
	Domain            string    `json:"domain"`
	Status            string    `json:"status"`
	Kind              string    `json:"kind"`
	RedirectToPrimary bool      `json:"redirect_to_primary"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateCompanyDomainRequest struct {
	Domain            string `json:"domain"`
	Kind              string `json:"kind"`
	Status            string `json:"status"`
	RedirectToPrimary bool   `json:"redirect_to_primary"`
}

type UpdateCompanyDomainRequest struct {
	Domain            string `json:"domain"`
	Kind              string `json:"kind"`
	Status            string `json:"status"`
	RedirectToPrimary bool   `json:"redirect_to_primary"`
}

func (r *CreateCompanyDomainRequest) Validate() validator.ValidationErrors {
	r.Domain = str.NormalizeDomain(r.Domain)
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "domain", r.Domain).Required().Apply(validator.MinLength(3))
	if !validCompanyDomainKind(r.Kind) {
		errs.Add("kind", messages.MsgInvalidFormat)
	}
	if !validCompanyDomainStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	if r.RedirectToPrimary && r.Kind == "primary" {
		errs.Add("redirect_to_primary", messages.MsgInvalidFormat)
	}
	return errs
}

func (r *UpdateCompanyDomainRequest) Validate() validator.ValidationErrors {
	r.Domain = str.NormalizeDomain(r.Domain)
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "domain", r.Domain).Required().Apply(validator.MinLength(3))
	if !validCompanyDomainKind(r.Kind) {
		errs.Add("kind", messages.MsgInvalidFormat)
	}
	if !validCompanyDomainStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	if r.RedirectToPrimary && r.Kind == "primary" {
		errs.Add("redirect_to_primary", messages.MsgInvalidFormat)
	}
	return errs
}

func validCompanyDomainKind(kind string) bool {
	return kind == "primary" || kind == "alias" || kind == "technical"
}

func validCompanyDomainStatus(status string) bool {
	return status == "active" || status == "inactive" || status == "pending_verification"
}

func (r *CompanyDomainResponse) FromModel(m CompanyDomain, companyPublicID string) {
	*r = CompanyDomainResponse{PublicID: m.PublicID, CompanyPublicID: companyPublicID, Domain: m.Domain, Status: m.Status, Kind: m.Kind, RedirectToPrimary: m.RedirectToPrimary, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}
