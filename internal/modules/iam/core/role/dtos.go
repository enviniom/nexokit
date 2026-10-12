package role

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"github.com/enviniom/nexokit/internal/platform/validator"
)

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

// FromModel projects a create, update, or list result. Permission selection and
// ordering belong to the caller; mapping preserves nil and empty slices.
func (r *RoleResponse) FromModel(model *IAMRole, permissions []string) {
	*r = RoleResponse{
		PublicID: model.PublicID, Name: model.Name, Slug: model.Slug,
		Description: model.Description, IsSystem: model.IsSystem,
		Permissions: permissions, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy, UpdatedBy: model.UpdatedBy,
	}
	if model.Company.PublicID != "" {
		id := model.Company.PublicID
		r.CompanyID = &id
	}
}

// FromViewModel retains the view contract: a company foreign key emits the
// preloaded public ID even when that ID is empty.
func (r *RoleResponse) FromViewModel(model *IAMRole, permissions []string) {
	r.FromModel(model, permissions)
	r.CompanyID = nil
	if model.CompanyID != nil {
		id := model.Company.PublicID
		r.CompanyID = &id
	}
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

func (r *RolePermissionResponse) FromModel(model *permission.IAMPermission, granted map[string]bool) {
	*r = RolePermissionResponse{
		PublicID: model.PublicID, Slug: model.Slug, Name: model.Name,
		Module: model.Module, Action: model.Action, Description: model.Description,
		IsSystem: model.IsSystem, DisplayOrder: model.DisplayOrder, Granted: granted[model.Slug],
	}
}

// BuildPermissionCatalog preserves catalog order and first-seen module order.
func BuildPermissionCatalog(catalog []permission.IAMPermission, granted map[string]bool) []RolePermissionGroupResponse {
	groups := []RolePermissionGroupResponse{}
	indices := map[string]int{}
	for i := range catalog {
		model := &catalog[i]
		index, exists := indices[model.Module]
		if !exists {
			groups = append(groups, RolePermissionGroupResponse{Module: model.Module})
			index = len(groups) - 1
			indices[model.Module] = index
		}
		var item RolePermissionResponse
		item.FromModel(model, granted)
		groups[index].Permissions = append(groups[index].Permissions, item)
	}
	return groups
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

func (r *RolePermissionAssignmentResponse) FromSelection(roleID string, permissions []string, catalog []RolePermissionGroupResponse) {
	*r = RolePermissionAssignmentResponse{RoleID: roleID, Permissions: permissions, Catalog: catalog}
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
