package goldenmod

import (
	"net/http"
	"strconv"

	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// GoldenmodHandler holds HTTP handlers for Goldenmod.
type GoldenmodHandler struct {
	service *GoldenmodService
}

// NewGoldenmodHandler creates a new handler instance.
func NewGoldenmodHandler(service *GoldenmodService) *GoldenmodHandler {
	return &GoldenmodHandler{service: service}
}

// Create handles POST /goldenmod.
func (h *GoldenmodHandler) Create(c *gin.Context) {
	var req CreateGoldenmodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Created(c, toGoldenmodResponse(m))
}

// Get handles GET /goldenmod/:id.
func (h *GoldenmodHandler) Get(c *gin.Context) {
	publicID := c.Param("id")
	m, err := h.service.Get(c.Request.Context(), publicID)
	if err != nil {
		response.NotFound(c, "Goldenmod not found")
		return
	}
	response.OK(c, toGoldenmodResponse(m))
}

// List handles GET /goldenmod with pagination.
func (h *GoldenmodHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 {
		perPage = 20
	}
	items, total, err := h.service.List(c.Request.Context(), page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Paginated(c, "Goldenmod list retrieved", toGoldenmodListResponse(items), page, perPage, total)
}

// Update handles PUT /goldenmod/:id.
func (h *GoldenmodHandler) Update(c *gin.Context) {
	publicID := c.Param("id")
	var req UpdateGoldenmodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	m, err := h.service.Update(c.Request.Context(), publicID, req)
	if err != nil {
		response.NotFound(c, "Goldenmod not found")
		return
	}
	response.OK(c, toGoldenmodResponse(m))
}

// Delete handles DELETE /goldenmod/:id.
func (h *GoldenmodHandler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), publicID); err != nil {
		response.NotFound(c, "Goldenmod not found")
		return
	}
	response.NoContent(c)
}

func toGoldenmodResponse(m *Goldenmod) GoldenmodResponse {
	return GoldenmodResponse{
		ID:        m.PublicID,
		Name:      m.Name,
		CompanyID: m.CompanyID,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toGoldenmodListResponse(items []Goldenmod) []GoldenmodResponse {
	out := make([]GoldenmodResponse, len(items))
	for i, m := range items {
		out[i] = toGoldenmodResponse(&m)
	}
	return out
}
