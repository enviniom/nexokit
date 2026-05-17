package users

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// UserResponse is the DTO for user read operations.
type UserResponse struct {
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

// CreateUserRequest is the DTO for creating a user.
type CreateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	RoleID    uint   `json:"role_id"`
	CompanyID *uint  `json:"company_id,omitempty"`
}

// Validate performs field-level validation for CreateUserRequest.
func (r CreateUserRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "email", r.Email).Required().Apply(validator.ValidEmail())
	validator.Field(errs, "password", r.Password).Required().Apply(validator.MinLength(8))
	if r.RoleID == 0 {
		errs.Add("role_id", messages.MsgRequired)
	}
	return errs
}

// UpdateUserRequest is the DTO for updating a user.
type UpdateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	RoleID    uint   `json:"role_id"`
	CompanyID *uint  `json:"company_id,omitempty"`
}

// Validate performs field-level validation for UpdateUserRequest.
func (r UpdateUserRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "email", r.Email).Required().Apply(validator.ValidEmail())
	if r.RoleID == 0 {
		errs.Add("role_id", messages.MsgRequired)
	}
	return errs
}

// ChangePasswordRequest is the DTO for changing a user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Validate performs field-level validation for ChangePasswordRequest.
func (r ChangePasswordRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "current_password", r.CurrentPassword).Required()
	validator.Field(errs, "new_password", r.NewPassword).Required().Apply(validator.MinLength(8))
	return errs
}

// UpdateStatusRequest is the DTO for toggling a user's active status.
type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// Validate performs field-level validation for UpdateStatusRequest.
func (r UpdateStatusRequest) Validate() response.ValidationErrors {
	return make(response.ValidationErrors)
}
