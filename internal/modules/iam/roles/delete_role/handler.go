package delete_role

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service Service }

func NewHandler(service Service) *Handler { return &Handler{service: service} }
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
	case errors.Is(err, core.ErrRoleProtected):
		return apperror.ErrForbidden
	case errors.Is(err, core.ErrRoleHasAssignedUsers):
		return apperror.Wrap(apperror.ErrUnprocessable, "El rol tiene usuarios asignados")
	default:
		return err
	}
}
