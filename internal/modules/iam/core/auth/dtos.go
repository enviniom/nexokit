package user

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/validator"
)

// LoginRequest is the DTO for login requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() validator.ValidationErrors {
	errs := make(validator.ValidationErrors)
	validator.Field(errs, "email", r.Email).Required().Apply(validator.ValidEmail())
	validator.Field(errs, "password", r.Password).Required()
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

func (r *AuthUserResponse) FromModel(model AuthUser) {
	*r = AuthUserResponse{
		PublicID: model.PublicID, Name: model.Name, Email: model.Email,
		IsActive: model.IsActive, RoleID: model.RoleID, RoleName: model.Role.Name,
		CompanyID: model.CompanyID, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy, UpdatedBy: model.UpdatedBy,
	}
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

// FromSession maps the sanitized auth-context projection, preserving /me's zero audit timestamps.
func (r *MeResponse) FromSession(view SessionView) {
	*r = MeResponse{
		AuthUserResponse: AuthUserResponse{
			PublicID: view.PublicID, Name: view.Name, Email: view.Email,
			IsActive: view.IsActive, RoleID: view.RoleID, RoleName: view.RoleName,
			CompanyID: view.CompanyID,
		},
		RoleSlug: view.RoleSlug, Permissions: view.Permissions,
	}
}
