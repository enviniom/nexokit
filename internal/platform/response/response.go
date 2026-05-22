package response

import (
	"net/http"
	"time"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all API responses.
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    any    `json:"meta"`
	Errors  any    `json:"errors"`
}

// ErrorResponse is the standard envelope for non-validation errors.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    any    `json:"meta"`
	Errors  any    `json:"errors"`
}

// ValidationErrorResponse is the standard envelope for field-keyed validation errors.
type ValidationErrorResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    any              `json:"data"`
	Meta    any              `json:"meta"`
	Errors  ValidationErrors `json:"errors"`
}

// PaginatedResponse is the standard envelope for paginated list responses.
type PaginatedResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Meta    any    `json:"meta"`
	Errors  any    `json:"errors"`
}

// PaginationMeta holds pagination information for list responses.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// FiltersMeta holds generic list filters reflected back to API consumers.
type FiltersMeta struct {
	Status      string `json:"status"`
	CreatedFrom string `json:"created_from"`
	CreatedTo   string `json:"created_to"`
	Sort        string `json:"sort"`
	Order       string `json:"order"`
	Search      string `json:"search"`
}

// ValidationErrors accumulates validation errors per field.
type ValidationErrors map[string][]string

// Add appends a message to the field's error list.
func (ve ValidationErrors) Add(field, message string) {
	ve[field] = append(ve[field], message)
}

// HasErrors returns true if there is at least one validation error.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Success returns a 200 OK success response.
func Success[T any](c *gin.Context, message string, data T) {
	c.JSON(http.StatusOK, APIResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    nil,
		Errors:  nil,
	})
}

// Created returns a 201 Created success response.
func Created[T any](c *gin.Context, message string, data T) {
	c.JSON(http.StatusCreated, APIResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    nil,
		Errors:  nil,
	})
}

// NoContent returns a 204 No Content response with no body.
func NoContent(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

// Error returns a generic error response with the given status code.
func Error(c *gin.Context, status int, message string, errs any) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Meta:    nil,
		Errors:  errs,
	})
}

// BadRequest returns a 400 Bad Request response.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message, nil)
}

// NotFound returns a 404 Not Found response.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, nil)
}

// Unauthorized returns a 401 Unauthorized response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, nil)
}

// Forbidden returns a 403 Forbidden response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, nil)
}

// Conflict returns a 409 Conflict response.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message, nil)
}

// TooManyRequests returns a 429 Too Many Requests response.
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, message, nil)
}

// InternalServerError returns a 500 Internal Server Error response.
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, nil)
}

// ValidationError returns a 422 Unprocessable Entity response with field errors.
func ValidationError(c *gin.Context, errs any) {
	validationErrs := ValidationErrors{}
	switch typed := errs.(type) {
	case ValidationErrors:
		validationErrs = typed
	case map[string][]string:
		validationErrs = ValidationErrors(typed)
	}
	if validationErrs == nil {
		validationErrs = ValidationErrors{}
	}
	c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
		Success: false,
		Message: messages.MsgValidationError,
		Data:    nil,
		Meta:    nil,
		Errors:  validationErrs,
	})
}

// Paginated returns a 200 OK response with pagination metadata.
func Paginated[T any](c *gin.Context, message string, data T, page, perPage int, total int64) {
	PaginatedWithFilters(c, message, data, query.ListParams{Pagination: query.PaginationParams{Page: page, PerPage: perPage}}, total)
}

// PaginatedWithFilters returns a 200 OK response with pagination and filter metadata.
func PaginatedWithFilters[T any](c *gin.Context, message string, data T, params query.ListParams, total int64) {
	page := params.Pagination.Page
	perPage := params.Pagination.PerPage
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	c.JSON(http.StatusOK, PaginatedResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta: map[string]any{
			"pagination": PaginationMeta{
				Page:       page,
				PerPage:    perPage,
				Total:      total,
				TotalPages: totalPages,
			},
			"filters": filtersMeta(params.Filters, params.Sort, params.Search),
		},
		Errors: nil,
	})
}

// HandleError maps application errors to standard API error responses.
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	Error(c, apperror.Status(err), apperror.PublicMessage(err, gin.Mode()), nil)
}

// RespondIfInvalid writes a 422 validation error response and returns true if errs has errors.
// Use in handlers to short-circuit when validation fails.
func RespondIfInvalid(c *gin.Context, errs ValidationErrors) bool {
	if len(errs) == 0 {
		return false
	}
	ValidationError(c, errs)
	return true
}

func filtersMeta(filters query.FilterParams, sort query.SortParams, search query.SearchParams) FiltersMeta {
	if sort.Sort == "" {
		sort.Sort = "created_at"
	}
	if sort.Order != "asc" && sort.Order != "desc" {
		sort.Order = "desc"
	}
	return FiltersMeta{
		Status:      filters.Status,
		CreatedFrom: formatDate(filters.CreatedFrom),
		CreatedTo:   formatDate(filters.CreatedTo),
		Sort:        sort.Sort,
		Order:       sort.Order,
		Search:      search.Query,
	}
}

func formatDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.DateOnly)
}
