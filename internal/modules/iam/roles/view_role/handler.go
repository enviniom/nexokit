package view_role

import (
	"errors"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
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
	item, err := h.service.View(tc, c.Param("id"))
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
	default:
		return err
	}
}
