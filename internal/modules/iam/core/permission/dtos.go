package permission

import (
	"github.com/enviniom/nexokit/internal/platform/validator"
	"time"
)

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

// FromModel maps the permission's public fields, replacing any previous response.
func (r *PermissionResponse) FromModel(model IAMPermission) {
	*r = PermissionResponse{
		PublicID: model.PublicID, Slug: model.Slug, Name: model.Name,
		Module: model.Module, Action: model.Action, Description: model.Description,
		IsSystem: model.IsSystem, DisplayOrder: model.DisplayOrder,
		CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy, UpdatedBy: model.UpdatedBy,
	}
}
