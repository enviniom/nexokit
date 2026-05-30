package update_permission

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
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

// Update updates a non-system permission.
func (h *Handler) Update(c *gin.Context) {
	var req core.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	permission, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			response.HandleError(c, apperror.ErrNotFound)
			return
		}
		if errors.Is(err, core.ErrConflict) || errors.Is(err, core.ErrSystemImmutable) {
			response.HandleError(c, apperror.ErrConflict)
			return
		}
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, permission)
}
