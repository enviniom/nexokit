package roles

import (
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the roles module.
type Handler struct {
	service Service
}

// NewHandler creates a new roles handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List returns paginated roles with filter metadata.
func (h *Handler) List(c *gin.Context) {
	params := query.ListFromGin(c)
	tc := h.tenantContext(c)

	roles, total, err := h.service.List(tc, params)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.PaginatedWithFilters(c, messages.MsgSuccess, roles, params, total)
}

// GetByPublicID returns a single role by its public ID.
func (h *Handler) GetByPublicID(c *gin.Context) {
	publicID := c.Param("id")
	role, err := h.service.GetByPublicID(h.tenantContext(c), publicID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, role)
}

// GetPermissionCatalog returns the full permission catalog annotated for a role.
func (h *Handler) GetPermissionCatalog(c *gin.Context) {
	publicID := c.Param("id")
	catalog, err := h.service.GetPermissionCatalog(h.tenantContext(c), publicID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, catalog)
}

// AssignPermissions replaces a role's permission assignments by slug.
func (h *Handler) AssignPermissions(c *gin.Context) {
	publicID := c.Param("id")
	var req AssignRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	actorPermissions := permissionSlugsFromContext(c)
	if !containsPermission(actorPermissions, "roles.assign_permissions") {
		response.Forbidden(c, messages.MsgForbidden)
		return
	}
	result, err := h.service.AssignPermissions(h.tenantContext(c), publicID, req, actorPermissions)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, result)
}

// Create creates a new role.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	role, err := h.service.Create(h.tenantContext(c), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, messages.MsgSuccess, role)
}

// Update updates an existing role.
func (h *Handler) Update(c *gin.Context) {
	publicID := c.Param("id")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	role, err := h.service.Update(h.tenantContext(c), publicID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, role)
}

// Delete deletes a role.
func (h *Handler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	if err := h.service.Delete(h.tenantContext(c), publicID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}

func permissionSlugsFromContext(c *gin.Context) []string {
	user, ok := authctx.FromGin(c)
	if !ok || user == nil {
		return nil
	}
	return user.Permissions
}

func containsPermission(items []string, slug string) bool {
	for _, item := range items {
		if item == "*" || item == slug {
			return true
		}
	}
	return false
}

func (h *Handler) tenantContext(c *gin.Context) tenant.TenantContext {
	if tc, ok := tenant.FromGin(c); ok {
		return tc
	}
	return tenant.NewRoot()
}
