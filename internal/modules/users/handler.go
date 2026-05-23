package users

import (
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the users module.
type Handler struct {
	service       Service
	actorProvider func(*gin.Context) string
}

// NewHandler creates a new users handler.
func NewHandler(service Service, actorProvider func(*gin.Context) string) *Handler {
	if actorProvider == nil {
		actorProvider = authctx.PublicIDFromGin
	}
	return &Handler{service: service, actorProvider: actorProvider}
}

// List returns paginated users.
func (h *Handler) List(c *gin.Context) {
	params := query.ListFromGin(c)

	tc := h.tenantContext(c)
	users, total, err := h.service.List(tc, params)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.PaginatedWithFilters(c, messages.MsgSuccess, users, params, total)
}

// GetByPublicID returns a single user by its public ID.
func (h *Handler) GetByPublicID(c *gin.Context) {
	publicID := c.Param("id")
	user, err := h.service.GetByPublicID(h.tenantContext(c), publicID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, user)
}

// Create creates a new user.
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	user, err := h.service.Create(h.tenantContext(c), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, messages.MsgSuccess, user)
}

// Update updates an existing user.
func (h *Handler) Update(c *gin.Context) {
	publicID := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	user, err := h.service.Update(h.tenantContext(c), publicID, h.actorPublicID(c), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, user)
}

// Delete soft-deletes a user by its public ID.
func (h *Handler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	if err := h.service.Delete(h.tenantContext(c), publicID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}

// ChangePassword handles PATCH /users/:id/password.
func (h *Handler) ChangePassword(c *gin.Context) {
	publicID := c.Param("id")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	if err := h.service.ChangePassword(h.tenantContext(c), publicID, h.actorPublicID(c), req); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}

// ChangeRole handles PATCH /users/:id/role.
func (h *Handler) ChangeRole(c *gin.Context) {
	publicID := c.Param("id")
	var req ChangeUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	user, err := h.service.ChangeRole(h.tenantContext(c), publicID, h.actorPublicID(c), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, user)
}

func (h *Handler) actorPublicID(c *gin.Context) string {
	if h.actorProvider == nil {
		return ""
	}
	return h.actorProvider(c)
}

func (h *Handler) tenantContext(c *gin.Context) tenant.TenantContext {
	if tc, ok := tenant.FromGin(c); ok {
		return tc
	}
	return tenant.NewRoot()
}

// ToggleStatus handles PATCH /users/:id/status.
func (h *Handler) ToggleStatus(c *gin.Context) {
	publicID := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	user, err := h.service.ToggleStatus(h.tenantContext(c), publicID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, user)
}
