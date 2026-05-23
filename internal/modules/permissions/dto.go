package permissions

import (
	"time"

	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

// PermissionResponse is the DTO for permission read operations.
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

// PermissionGroupResponse groups permissions by module for UI rendering.
type PermissionGroupResponse struct {
	Module      string               `json:"module"`
	Permissions []PermissionResponse `json:"permissions"`
}

// CreatePermissionRequest is the DTO for creating a non-system permission.
type CreatePermissionRequest struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Module       string `json:"module"`
	Action       string `json:"action"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

// Validate performs field-level validation for CreatePermissionRequest.
func (r CreatePermissionRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "slug", r.Slug).Required()
	validator.Field(errs, "name", r.Name).Required()
	validator.Field(errs, "module", r.Module).Required()
	validator.Field(errs, "action", r.Action).Required()
	return errs
}

// UpdatePermissionRequest is the DTO for updating a non-system permission.
type UpdatePermissionRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

// Validate performs field-level validation for UpdatePermissionRequest.
func (r UpdatePermissionRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required()
	return errs
}
