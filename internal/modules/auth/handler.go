package auth

import (
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the auth module.
type Handler struct {
	service Service
}

// NewHandler creates a new auth handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Login handles POST /auth/login.
func (h *Handler) Login(c *gin.Context) {
	// TODO(multitenancy): make login tenant-aware for company domains. When a
	// request comes from a company host, resolve that host to a company and reject
	// non-root users whose company_id does not match the resolved tenant, even if
	// they try to send a different tenant header.
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}

	result, err := h.service.Login(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, result)
}

// Refresh handles POST /auth/refresh.
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}

	result, err := h.service.Refresh(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, result)
}

// Logout handles POST /auth/logout.
func (h *Handler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	if err := h.service.Logout(req); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}

// Me returns the authenticated user from context without sensitive fields.
func (h *Handler) Me(c *gin.Context) {
	current, ok := authctx.FromGin(c)
	if !ok || current == nil {
		response.Unauthorized(c, messages.MsgUnauthorized)
		return
	}
	response.Success(c, messages.MsgSuccess, MeResponse{UserResponse: users.UserResponse{
		PublicID:  current.PublicID,
		Name:      current.Name,
		Email:     current.Email,
		IsActive:  current.IsActive,
		RoleID:    current.RoleID,
		RoleName:  current.Role,
		CompanyID: current.CompanyID,
	}, RoleSlug: current.RoleSlug, Permissions: current.Permissions})
}
