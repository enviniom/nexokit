package delete_user

import (
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler is the HTTP entry point for the delete_user use case.
type Handler struct{ service Service }

// NewHandler creates a delete_user handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle soft-deletes a user by public ID and responds with 204 No Content.
func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok {
		tc = tenant.NewRoot()
	}
	if err := h.service.Delete(tc, c.Param("id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}
