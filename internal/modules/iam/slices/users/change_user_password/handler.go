package change_user_password

import (
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler is the HTTP entry point for the change_user_password use case.
type Handler struct{ service Service }

// NewHandler creates a change_user_password handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle changes a user's password from a JSON request body.
func (h *Handler) Handle(c *gin.Context) {
	var req iamuser.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	tc, ok := tenant.FromGin(c)
	if !ok {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if err := h.service.Change(tc, c.Param("id"), authctx.PublicIDFromGin(c), req); err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}
