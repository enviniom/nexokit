package company

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/companies/core/company_domain"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

type CompanyResponse struct {
	PublicID  string                                 `json:"public_id"`
	Name      string                                 `json:"name"`
	Slug      string                                 `json:"slug"`
	Status    string                                 `json:"status"`
	Domains   []company_domain.CompanyDomainResponse `json:"domains,omitempty"`
	CreatedAt time.Time                              `json:"created_at"`
	UpdatedAt time.Time                              `json:"updated_at"`
	CreatedBy *uint                                  `json:"created_by,omitempty"`
	UpdatedBy *uint                                  `json:"updated_by,omitempty"`
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

func (r *CreateCompanyRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.MinLength(2))
	if r.Status != "" && !validStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	return errs
}

func (r *UpdateCompanyRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.MinLength(2))
	if !validStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	return errs
}

func validStatus(status string) bool {
	return status == "active" || status == "inactive"
}

// FromModel maps the company projection; callers explicitly select domain inclusion.
func (r *CompanyResponse) FromModel(m Company, includeDomains bool) {
	*r = CompanyResponse{PublicID: m.PublicID, Name: m.Name, Slug: m.Slug, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, CreatedBy: m.CreatedBy, UpdatedBy: m.UpdatedBy}
	if includeDomains && len(m.Domains) > 0 {
		r.Domains = make([]company_domain.CompanyDomainResponse, len(m.Domains))
		for i := range m.Domains {
			r.Domains[i].FromModel(m.Domains[i], m.PublicID)
		}
	}
}
