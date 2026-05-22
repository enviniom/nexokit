package roles

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// RoleResponse is the DTO for role read operations.
type RoleResponse struct {
	PublicID    string    `json:"public_id"`
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

// RolePermissionResponse is a permission DTO annotated with role assignment state.
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

// RolePermissionGroupResponse groups role permission catalog entries by module.
type RolePermissionGroupResponse struct {
	Module      string                   `json:"module"`
	Permissions []RolePermissionResponse `json:"permissions"`
}

// AssignRolePermissionsRequest replaces a role's permission assignments by slug.
type AssignRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

// Validate performs field-level validation for AssignRolePermissionsRequest.
func (r AssignRolePermissionsRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	if r.Permissions == nil {
		errs.Add("permissions", "permissions is required")
	}
	return errs
}

// RolePermissionAssignmentResponse returns replacement result and refreshed catalog.
type RolePermissionAssignmentResponse struct {
	RoleID      string                        `json:"role_id"`
	Permissions []string                      `json:"permissions"`
	Catalog     []RolePermissionGroupResponse `json:"catalog"`
}

// CreateRoleRequest is the DTO for creating a role.
type CreateRoleRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// Validate performs field-level validation for CreateRoleRequest.
func (r CreateRoleRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.ValidSlug())
	return errs
}

// UpdateRoleRequest is the DTO for updating a role.
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// Validate performs field-level validation for UpdateRoleRequest.
func (r UpdateRoleRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.ValidSlug())
	return errs
}
