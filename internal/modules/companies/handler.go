package companies

import (
	"errors"

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
	params := query.ListFromGin(c)
	req := ListCompaniesRequest{ListParams: params, IncludeInactive: c.Query("include_inactive") == "true"}
	companies, total, err := h.service.List(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.PaginatedWithFilters(c, messages.MsgSuccess, companies, params, total)
}

func (h *Handler) GetByPublicID(c *gin.Context) {
	company, err := h.service.GetByPublicID(c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
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
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) ListDomains(c *gin.Context) {
	domains, err := h.service.ListDomains(c.Param("id"))
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, domains)
}

func (h *Handler) CreateDomain(c *gin.Context) {
	var req CreateCompanyDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	domain, err := h.service.CreateDomain(c.Param("id"), req)
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.Created(c, messages.MsgSuccess, domain)
}

func (h *Handler) UpdateDomain(c *gin.Context) {
	var req UpdateCompanyDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, messages.MsgBadRequest)
		return
	}
	if response.RespondIfInvalid(c, req.Validate()) {
		return
	}
	domain, err := h.service.UpdateDomain(c.Param("id"), c.Param("domain_id"), req)
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.Success(c, messages.MsgSuccess, domain)
}

func (h *Handler) respondError(c *gin.Context, err error) {
	if errors.Is(err, ErrDuplicateSlug) {
		errs := make(response.ValidationErrors)
		errs.Add("slug", messages.MsgConflict)
		response.ValidationError(c, errs)
		return
	}
	if errors.Is(err, ErrDuplicateCompanyDomain) {
		errs := make(response.ValidationErrors)
		errs.Add("domain", messages.MsgConflict)
		response.ValidationError(c, errs)
		return
	}
	if errors.Is(err, ErrActivePrimaryDomainExists) {
		errs := make(response.ValidationErrors)
		errs.Add("kind", messages.MsgConflict)
		response.ValidationError(c, errs)
		return
	}
	if errors.Is(err, ErrCompanyDomainDoesNotBelong) {
		response.NotFound(c, messages.MsgNotFound)
		return
	}
	response.HandleError(c, err)
}
