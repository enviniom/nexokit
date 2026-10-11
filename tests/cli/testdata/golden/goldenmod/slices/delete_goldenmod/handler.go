package delete_goldenmod

import (
	"github.com/enviniom/nexokit/internal/modules/goldenmod/core"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type Handler struct{ repo core.DeleteRepository }

func NewHandler(repo core.DeleteRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(c *gin.Context) {
	tc, ok := tenant.FromGin(c)
	if !ok || tc.IsRootScope || tc.CompanyID == 0 {
		response.HandleError(c, core.ErrTenantRequired)
		return
	}
	if err := h.repo.Delete(c.Request.Context(), tc, c.Param("id")); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}
