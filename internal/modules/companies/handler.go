package companies

import (
	"errors"
	"net/http"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// Handler handles company HTTP requests.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	req := ListCompaniesRequest{Pagination: query.PaginationFromGin(c), IncludeInactive: c.Query("include_inactive") == "true", Status: c.Query("status")}
	companies, total, err := h.service.List(req)
	if err != nil {
		response.InternalServerError(c, messages.MsgInternalError)
		return
	}
	response.Paginated(c, messages.MsgSuccess, companies, req.Page, req.PerPage, total)
}

func (h *Handler) GetByPublicID(c *gin.Context) {
	company, err := h.service.GetByPublicID(c.Param("id"))
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, company)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	company, err := h.service.Create(req)
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.Created(c, messages.MsgSuccess, company)
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	company, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, company)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		h.respondError(c, err)
		return
	}
	response.Success[any](c, messages.MsgSuccess, nil)
}

func (h *Handler) respondError(c *gin.Context, err error) {
	if errors.Is(err, ErrDuplicateSlug) {
		errs := make(response.ValidationErrors)
		errs.Add("slug", messages.MsgConflict)
		response.ValidationError(c, errs)
		return
	}
	switch apperror.Status(err) {
	case http.StatusNotFound:
		response.NotFound(c, messages.MsgNotFound)
	case http.StatusForbidden:
		response.Forbidden(c, messages.MsgForbidden)
	default:
		response.InternalServerError(c, messages.MsgInternalError)
	}
}
