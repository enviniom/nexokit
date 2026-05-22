package permissions

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the permissions module.
type Handler struct {
	service Service
}

// NewHandler creates a new permissions handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List returns permissions grouped by module.
func (h *Handler) List(c *gin.Context) {
	groups, err := h.service.ListGrouped()
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, groups)
}

// ListPaginated returns paginated permissions with filter metadata.
func (h *Handler) ListPaginated(c *gin.Context) {
	params := query.ListFromGin(c)
	permissions, total, err := h.service.List(params)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.PaginatedWithFilters(c, messages.MsgSuccess, permissions, params, total)
}

// GetByPublicID returns a single permission by public ID.
func (h *Handler) GetByPublicID(c *gin.Context) {
	permission, err := h.service.GetByPublicID(c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, permission)
}

// Create creates a non-system permission.
func (h *Handler) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	permission, err := h.service.Create(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, messages.MsgSuccess, permission)
}

// Update updates a non-system permission.
func (h *Handler) Update(c *gin.Context) {
	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	permission, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, permission)
}

// Delete deletes a non-system permission.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}
