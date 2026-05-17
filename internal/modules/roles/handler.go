package roles

import (
	"net/http"
	"strconv"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the roles module.
type Handler struct {
	service Service
}

// NewHandler creates a new roles handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List returns paginated roles.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 {
		perPage = 20
	}

	roles, total, err := h.service.List(page, perPage)
	if err != nil {
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Paginated(c, messages.MsgSuccess, roles, page, perPage, total)
}

// GetByPublicID returns a single role by its public ID.
func (h *Handler) GetByPublicID(c *gin.Context) {
	publicID := c.Param("id")
	role, err := h.service.GetByPublicID(publicID)
	if err != nil {
		if apperror.Status(err) == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success(c, messages.MsgSuccess, role)
}

// Create creates a new role.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		response.ValidationError(c, errs)
		return
	}
	role, err := h.service.Create(req)
	if err != nil {
		status := apperror.Status(err)
		if status == http.StatusConflict {
			response.Conflict(c, messages.MsgConflict)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Created(c, messages.MsgSuccess, role)
}

// Update updates an existing role.
func (h *Handler) Update(c *gin.Context) {
	publicID := c.Param("id")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		response.ValidationError(c, errs)
		return
	}
	role, err := h.service.Update(publicID, req)
	if err != nil {
		status := apperror.Status(err)
		if status == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		if status == http.StatusConflict {
			response.Conflict(c, messages.MsgConflict)
			return
		}
		if status == http.StatusForbidden {
			response.Forbidden(c, messages.MsgForbidden)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success(c, messages.MsgSuccess, role)
}

// Delete deletes a role.
func (h *Handler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	if err := h.service.Delete(publicID); err != nil {
		status := apperror.Status(err)
		if status == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		if status == http.StatusForbidden {
			response.Forbidden(c, messages.MsgForbidden)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}
