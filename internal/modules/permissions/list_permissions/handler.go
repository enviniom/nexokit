package list_permissions

import (
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

// List returns permissions grouped by module.
func (h *Handler) List(c *gin.Context) {
	groups, err := h.service.ListGrouped()
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, groups)
}
