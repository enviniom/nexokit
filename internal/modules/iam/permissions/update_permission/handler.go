package update_permission

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) Handle(c *gin.Context) {
	var req core.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}

	item, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		response.HandleError(c, mapServiceError(err))
		return
	}
	response.Success(c, messages.MsgSuccess, item)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return apperror.ErrNotFound
	case errors.Is(err, core.ErrConflict):
		return apperror.ErrConflict
	case errors.Is(err, core.ErrSystemImmutable):
		return apperror.ErrForbidden
	default:
		return err
	}
}
