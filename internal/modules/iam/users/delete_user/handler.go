package delete_user

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
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
		response.HandleError(c, mapServiceError(err))
		return
	}
	response.NoContent(c)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return apperror.ErrNotFound
	default:
		return err
	}
}
