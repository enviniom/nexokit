package goldenmod

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// CreateGoldenmodRequest is the payload for creating a Goldenmod.
type CreateGoldenmodRequest struct {
	Name      string `json:"name"`
	CompanyID uint   `json:"company_id"`
}

// Validate performs field-level validation for CreateGoldenmodRequest.
func (r CreateGoldenmodRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	if r.CompanyID == 0 {
		errs.Add("company_id", messages.MsgRequired)
	}
	return errs
}

// UpdateGoldenmodRequest is the payload for updating a Goldenmod.
type UpdateGoldenmodRequest struct {
	Name      string `json:"name"`
	CompanyID uint   `json:"company_id"`
}

// Validate performs field-level validation for UpdateGoldenmodRequest.
func (r UpdateGoldenmodRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	if r.CompanyID == 0 {
		errs.Add("company_id", messages.MsgRequired)
	}
	return errs
}

// GoldenmodResponse is the JSON representation returned by the API.
type GoldenmodResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CompanyID uint   `json:"company_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
