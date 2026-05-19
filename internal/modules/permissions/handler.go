package permissions

import (
	"net/http"
	"strconv"

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

// List returns permissions grouped by module.
func (h *Handler) List(c *gin.Context) {
	groups, err := h.service.ListGrouped()
	if err != nil {
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success(c, messages.MsgSuccess, groups)
}

// ListPaginated returns paginated permissions.
func (h *Handler) ListPaginated(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 {
		perPage = 20
	}
	permissions, total, err := h.service.List(page, perPage)
	if err != nil {
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Paginated(c, messages.MsgSuccess, permissions, page, perPage, total)
}

// GetByPublicID returns a single permission by public ID.
func (h *Handler) GetByPublicID(c *gin.Context) {
	permission, err := h.service.GetByPublicID(c.Param("id"))
	if err != nil {
		if apperror.Status(err) == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success(c, messages.MsgSuccess, permission)
}

// Create creates a non-system permission.
func (h *Handler) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		response.ValidationError(c, errs)
		return
	}
	permission, err := h.service.Create(req)
	if err != nil {
		writePermissionError(c, err)
		return
	}
	response.Created(c, messages.MsgSuccess, permission)
}

// Update updates a non-system permission.
func (h *Handler) Update(c *gin.Context) {
	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		response.ValidationError(c, errs)
		return
	}
	permission, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		writePermissionError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, permission)
}

// Delete deletes a non-system permission.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		writePermissionError(c, err)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}

func writePermissionError(c *gin.Context, err error) {
	switch apperror.Status(err) {
	case http.StatusNotFound:
		response.NotFound(c, messages.MsgNotFound)
	case http.StatusConflict:
		response.Conflict(c, messages.MsgConflict)
	case http.StatusForbidden:
		response.Forbidden(c, messages.MsgForbidden)
	case http.StatusBadRequest:
		response.BadRequest(c, messages.MsgBadRequest)
	default:
		response.InternalServerError(c, messages.MsgInternalError)
	}
}
