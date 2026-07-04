package assign_role_to_user

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler is the HTTP entry point for the assign_role_to_user use case.
type Handler struct{ service Service }

// NewHandler creates an assign_role_to_user handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle assigns a new role to a user by public ID.
func (h *Handler) Handle(c *gin.Context) {
	var req core.ChangeUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}

	tc, ok := tenant.FromGin(c)
	if !ok {
		tc = tenant.NewRoot()
	}

	data, err := h.service.ChangeRole(tc, c.Param("id"), authctx.PublicIDFromGin(c), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, data)
}
