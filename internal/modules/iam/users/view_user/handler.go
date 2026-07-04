package view_user

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler is the HTTP entry point for the view_user use case.
type Handler struct{ service Service }

// NewHandler creates a view_user handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle responds with a single user by public ID.
func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok {
		tc = tenant.NewRoot()
	}
	data, err := h.service.GetByPublicID(tc, c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, data)
}
