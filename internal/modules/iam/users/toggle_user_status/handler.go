package toggle_user_status

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler is the HTTP entry point for the toggle_user_status use case.
type Handler struct{ service Service }

// NewHandler creates a toggle_user_status handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle toggles the active status of a user by public ID.
func (h *Handler) Handle(c *gin.Context) {
	var req core.UpdateStatusRequest
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

	data, err := h.service.Toggle(tc, c.Param("id"), req)
	if err != nil {
		response.HandleError(c, mapServiceError(err))
		return
	}
	response.Success(c, messages.MsgSuccess, data)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return apperror.ErrNotFound
	default:
		return err
	}
}
