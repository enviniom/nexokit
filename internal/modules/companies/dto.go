package companies

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// CompanyResponse is the external DTO for company read operations.
type CompanyResponse struct {
	PublicID  string    `json:"public_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Domain    *string   `json:"domain,omitempty"`
	Subdomain *string   `json:"subdomain,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy *uint     `json:"created_by,omitempty"`
	UpdatedBy *uint     `json:"updated_by,omitempty"`
}

// ListCompaniesRequest carries list filters and pagination.
type ListCompaniesRequest struct {
	query.Pagination
	IncludeInactive bool
	Status          string
}

// CreateCompanyRequest is the DTO for creating a company.
type CreateCompanyRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Domain    *string `json:"domain,omitempty"`
	Subdomain *string `json:"subdomain,omitempty"`
	Status    string  `json:"status,omitempty"`
}

// UpdateCompanyRequest is the DTO for updating a company.
type UpdateCompanyRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Domain    *string `json:"domain,omitempty"`
	Subdomain *string `json:"subdomain,omitempty"`
	Status    string  `json:"status"`
}

func (r CreateCompanyRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.MinLength(2))
	if r.Status != "" && !validStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	return errs
}

func (r UpdateCompanyRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.MinLength(2))
	if !validStatus(r.Status) {
		errs.Add("status", messages.MsgInvalidFormat)
	}
	return errs
}

func validStatus(status string) bool {
	return status == CompanyStatusActive || status == CompanyStatusInactive
}
