package goldenmod

import (
	"strings"
	"time"

	"github.com/enviniom/nexokit/internal/platform/validator"
)

// CreateGoldenmodRequest contains only client-editable fields.
// Company scope, when enabled, comes exclusively from trusted middleware.
type CreateGoldenmodRequest struct {
	Name string `json:"name"`
}

func (r *CreateGoldenmodRequest) Validate() validator.ValidationErrors {
	r.Name = strings.TrimSpace(r.Name)
	errs := validator.ValidationErrors{}
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	return errs
}

type UpdateGoldenmodRequest struct {
	Name string `json:"name"`
}

func (r *UpdateGoldenmodRequest) Validate() validator.ValidationErrors {
	r.Name = strings.TrimSpace(r.Name)
	errs := validator.ValidationErrors{}
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	return errs
}

// GoldenmodResponse never exposes persistence keys or audit identities.
type GoldenmodResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (r *GoldenmodResponse) FromModel(m Goldenmod) {
	r.ID = m.PublicID
	r.Name = m.Name
	r.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	r.UpdatedAt = m.UpdatedAt.Format(time.RFC3339)
}
