package core

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

type RoleResponse struct {
	PublicID    string    `json:"public_id"`
	CompanyID   *string   `json:"company_id,omitempty"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   *uint     `json:"created_by,omitempty"`
	UpdatedBy   *uint     `json:"updated_by,omitempty"`
}

type RolePermissionResponse struct {
	PublicID     string `json:"public_id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Module       string `json:"module"`
	Action       string `json:"action"`
	Description  string `json:"description"`
	IsSystem     bool   `json:"is_system"`
	DisplayOrder int    `json:"display_order"`
	Granted      bool   `json:"granted"`
}

type RolePermissionGroupResponse struct {
	Module      string                   `json:"module"`
	Permissions []RolePermissionResponse `json:"permissions"`
}

type AssignRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (r AssignRolePermissionsRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	if r.Permissions == nil {
		err.Add("permissions", "permissions is required")
	}
	return err
}

type RolePermissionAssignmentResponse struct {
	RoleID      string                        `json:"role_id"`
	Permissions []string                      `json:"permissions"`
	Catalog     []RolePermissionGroupResponse `json:"catalog"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func (r CreateRoleRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	validator.Field(err, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(err, "slug", r.Slug).Required().Apply(validator.ValidSlug())
	return err
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func (r UpdateRoleRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	validator.Field(err, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(err, "slug", r.Slug).Required().Apply(validator.ValidSlug())
	return err
}

type PermissionResponse struct {
	PublicID     string    `json:"public_id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Module       string    `json:"module"`
	Action       string    `json:"action"`
	Description  string    `json:"description"`
	IsSystem     bool      `json:"is_system"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedBy    *uint     `json:"created_by,omitempty"`
	UpdatedBy    *uint     `json:"updated_by,omitempty"`
}

type PermissionGroupResponse struct {
	Module      string               `json:"module"`
	Permissions []PermissionResponse `json:"permissions"`
}

type UpdatePermissionRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

func (r UpdatePermissionRequest) Validate() validator.ValidationErrors {
	err := make(validator.ValidationErrors)
	validator.Field(err, "name", r.Name).Required()
	return err
}
