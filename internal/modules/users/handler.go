package users

import (
	"net/http"
	"strconv"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the users module.
type Handler struct {
	service Service
}

// NewHandler creates a new users handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List returns paginated users.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 {
		perPage = 20
	}

	users, total, err := h.service.List(page, perPage)
	if err != nil {
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Paginated(c, messages.MsgSuccess, users, page, perPage, total)
}

// GetByPublicID returns a single user by its public ID.
func (h *Handler) GetByPublicID(c *gin.Context) {
	publicID := c.Param("id")
	user, err := h.service.GetByPublicID(publicID)
	if err != nil {
		if apperror.Status(err) == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success(c, messages.MsgSuccess, user)
}

// Create creates a new user.
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		response.ValidationError(c, errs)
		return
	}
	user, err := h.service.Create(req)
	if err != nil {
		status := apperror.Status(err)
		if status == http.StatusConflict {
			response.Conflict(c, messages.MsgConflict)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Created(c, messages.MsgSuccess, user)
}

// Update updates an existing user.
func (h *Handler) Update(c *gin.Context) {
	publicID := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		response.ValidationError(c, errs)
		return
	}
	// TODO(PR3): pass authenticated user's public ID as actorPublicID.
	user, err := h.service.Update(publicID, "", req)
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
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success(c, messages.MsgSuccess, user)
}

// Delete soft-deletes a user by its public ID.
func (h *Handler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	if err := h.service.Delete(publicID); err != nil {
		status := apperror.Status(err)
		if status == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}

// ChangePassword handles PATCH /users/:id/password.
func (h *Handler) ChangePassword(c *gin.Context) {
	publicID := c.Param("id")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		response.ValidationError(c, errs)
		return
	}
	// TODO(PR3): pass authenticated user's public ID as actorPublicID.
	if err := h.service.ChangePassword(publicID, "", req); err != nil {
		status := apperror.Status(err)
		if status == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		if status == http.StatusUnauthorized {
			response.Unauthorized(c, messages.MsgUnauthorized)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}

// ToggleStatus handles PATCH /users/:id/status.
func (h *Handler) ToggleStatus(c *gin.Context) {
	publicID := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.service.ToggleStatus(publicID, req)
	if err != nil {
		status := apperror.Status(err)
		if status == http.StatusNotFound {
			response.NotFound(c, messages.MsgNotFound)
			return
		}
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Success(c, messages.MsgSuccess, user)
}
