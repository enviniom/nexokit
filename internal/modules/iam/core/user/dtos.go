package user

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

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

// FromModel projects the user's public fields and the preloaded role name.
func (r *UserResponse) FromModel(model *IAMUser) {
	*r = UserResponse{
		PublicID: model.PublicID, Name: model.Name, Email: model.Email,
		IsActive: model.IsActive, RoleID: model.RoleID, RoleName: model.Role.Name,
		CompanyID: model.CompanyID, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy, UpdatedBy: model.UpdatedBy,
	}
}

type CreateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	RoleID    uint   `json:"role_id"`
	CompanyID *uint  `json:"company_id,omitempty"`
}

func (r CreateUserRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	validator.Field(err, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(err, "email", r.Email).Required().Apply(validator.ValidEmail())
	validator.Field(err, "password", r.Password).Required().Apply(validator.MinLength(8))
	if r.RoleID == 0 {
		err.Add("role_id", messages.MsgRequired)
	}
	return err
}

type UpdateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	CompanyID *uint  `json:"company_id,omitempty"`
}

func (r UpdateUserRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	validator.Field(err, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(err, "email", r.Email).Required().Apply(validator.ValidEmail())
	return err
}

type ChangeUserRoleRequest struct {
	RoleID string `json:"role_id"`
}

func (r ChangeUserRoleRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	validator.Field(err, "role_id", r.RoleID).Required()
	return err
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (r ChangePasswordRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	validator.Field(err, "current_password", r.CurrentPassword).Required()
	validator.Field(err, "new_password", r.NewPassword).Required().Apply(validator.MinLength(8))
	return err
}

type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func (r UpdateStatusRequest) Validate() validator.ValidationErrors {
	return make(validator.ValidationErrors)
}
