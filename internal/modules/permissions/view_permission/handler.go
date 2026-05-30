package view_permission

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

// GetByPublicID returns a single permission by public ID.
func (h *Handler) GetByPublicID(c *gin.Context) {
	permission, err := h.service.GetByPublicID(c.Param("id"))
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			response.HandleError(c, apperror.ErrNotFound)
			return
		}
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, permission)
}
