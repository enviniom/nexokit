package response

import (
	"net/http"

	"github.com/enviniom/nexokit/internal/platform/messages"
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

// PaginationMeta holds pagination information for list responses.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
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

// Error returns a generic error response with the given status code.
func Error(c *gin.Context, status int, message string, errs any) {
	c.JSON(status, APIResponse[any]{
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

// InternalServerError returns a 500 Internal Server Error response.
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, nil)
}

// ValidationError returns a 422 Unprocessable Entity response with field errors.
func ValidationError(c *gin.Context, errs any) {
	Error(c, http.StatusUnprocessableEntity, messages.MsgValidationError, errs)
}

// Paginated returns a 200 OK response with pagination metadata.
func Paginated[T any](c *gin.Context, message string, data T, page, perPage int, total int64) {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	c.JSON(http.StatusOK, APIResponse[T]{
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
		},
		Errors: nil,
	})
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
