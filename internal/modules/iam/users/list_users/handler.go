package list_users

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// Handler is the HTTP entry point for the list_users use case.
type Handler struct{ service Service }

// NewHandler creates a list_users handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Handle responds with a paginated, filtered list of users.
func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok {
		tc = tenant.NewRoot()
	}
	params := query.ListFromGin(c)
	data, total, err := h.service.List(tc, params)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.PaginatedWithFilters(c, messages.MsgSuccess, data, params, total)
}
