package core

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// LoginRequest is the DTO for login requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "email", r.Email).Required().Apply(validator.ValidEmail())
	validator.Field(errs, "password", r.Password).Required()
	return errs
}

// RefreshRequest is the DTO for refresh and logout requests.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r RefreshRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "refresh_token", r.RefreshToken).Required()
	return errs
}

// AuthUserResponse is the auth-scoped user DTO.
type AuthUserResponse struct {
	PublicID  string    `json:"public_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	RoleID    uint      `json:"role_id"`
	RoleName  string    `json:"role_name"`
	CompanyID *uint     `json:"company_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy *uint     `json:"created_by,omitempty"`
	UpdatedBy *uint     `json:"updated_by,omitempty"`
}

// TokenPairResponse returns a rotated token pair.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse returns tokens and a sanitized user DTO.
type LoginResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	User         AuthUserResponse `json:"user"`
}

// MeResponse returns the authenticated user with role and permission metadata.
type MeResponse struct {
	AuthUserResponse
	RoleSlug    string   `json:"role_slug"`
	Permissions []string `json:"permissions"`
}
