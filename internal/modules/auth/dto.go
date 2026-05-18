package auth

import (
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// LoginRequest is the DTO for login requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate performs field-level validation for LoginRequest.
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

// Validate performs field-level validation for RefreshRequest.
func (r RefreshRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "refresh_token", r.RefreshToken).Required()
	return errs
}

// TokenPairResponse returns a rotated token pair.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse returns tokens and a sanitized user DTO.
type LoginResponse struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	User         users.UserResponse `json:"user"`
}
