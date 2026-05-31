package core

import (
	"strings"
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

type CompanyResponse struct {
	PublicID  string                  `json:"public_id"`
	Name      string                  `json:"name"`
	Slug      string                  `json:"slug"`
	Status    string                  `json:"status"`
	Domains   []CompanyDomainResponse `json:"domains,omitempty"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	CreatedBy *uint                   `json:"created_by,omitempty"`
	UpdatedBy *uint                   `json:"updated_by,omitempty"`
}

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

type ListCompaniesRequest struct {
	query.ListParams
	IncludeInactive bool
}

type CreateCompanyRequest struct {
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Status string `json:"status,omitempty"`
}

type UpdateCompanyRequest struct {
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Status string `json:"status"`
}

func (r CreateCompanyRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.MinLength(2))
	if r.Status != "" && !validStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	return errs
}

func (r UpdateCompanyRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.MinLength(2))
	if !validStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	return errs
}

func (r CreateCompanyDomainRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	domain := normalizeCompanyDomain(r.Domain)
	validator.Field(errs, "domain", domain).Required().Apply(validator.MinLength(3))
	if !validCompanyDomainKind(r.Kind) {
		errs.Add("kind", messages.MsgInvalidFormat)
	}
	if !validCompanyDomainStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	if r.RedirectToPrimary && r.Kind == CompanyDomainKindPrimary {
		errs.Add("redirect_to_primary", messages.MsgInvalidFormat)
	}
	return errs
}

func (r UpdateCompanyDomainRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	domain := normalizeCompanyDomain(r.Domain)
	validator.Field(errs, "domain", domain).Required().Apply(validator.MinLength(3))
	if !validCompanyDomainKind(r.Kind) {
		errs.Add("kind", messages.MsgInvalidFormat)
	}
	if !validCompanyDomainStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	if r.RedirectToPrimary && r.Kind == CompanyDomainKindPrimary {
		errs.Add("redirect_to_primary", messages.MsgInvalidFormat)
	}
	return errs
}

func normalizeCompanyDomain(input string) string {
	normalized := strings.TrimSpace(strings.ToLower(input))
	return strings.TrimSuffix(normalized, ".")
}

func validStatus(status string) bool {
	return status == CompanyStatusActive || status == CompanyStatusInactive
}

func validCompanyDomainKind(kind string) bool {
	return kind == CompanyDomainKindPrimary || kind == CompanyDomainKindAlias || kind == CompanyDomainKindTechnical
}

func validCompanyDomainStatus(status string) bool {
	return status == CompanyDomainStatusActive || status == CompanyDomainStatusInactive || status == CompanyDomainStatusPendingVerification
}
