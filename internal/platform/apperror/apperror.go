package apperror

import (
	"errors"
	"net/http"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

// AppError represents an application-level error with a user-facing message.
type AppError struct {
	Err     error
	Message string
	Cause   error
}

func (ae *AppError) Error() string {
	if ae.Err != nil {
		return ae.Err.Error()
	}
	return ae.Message
}

func (ae *AppError) Unwrap() error {
	return ae.Cause
}

// Is implements error matching for errors.Is.
func (ae *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		if ae == t {
			return true
		}
		if ae.Err == nil && t.Err == nil && ae.Message == t.Message {
			return true
		}
	}
	if ae.Err != nil {
		return errors.Is(ae.Err, target)
	}
	return false
}

// Sentinel errors for common application states.
var (
	ErrNotFound        = &AppError{Message: messages.MsgNotFound}
	ErrForbidden       = &AppError{Message: messages.MsgForbidden}
	ErrUnauthorized    = &AppError{Message: messages.MsgUnauthorized}
	ErrConflict        = &AppError{Message: messages.MsgConflict}
	ErrBadRequest      = &AppError{Message: messages.MsgBadRequest}
	ErrTooManyRequests = &AppError{Message: messages.MsgTooManyRequests}
	ErrValidation      = &AppError{Message: messages.MsgValidationError}
	ErrInternal        = &AppError{Message: messages.MsgInternalError}
)

// Wrap creates a new AppError wrapping an existing error.
func Wrap(err error, message string, cause ...error) *AppError {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}
	return &AppError{Err: err, Message: message, Cause: c}
}

// Status returns the appropriate HTTP status code for an error.
func Status(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrForbidden) {
		return http.StatusForbidden
	}
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict
	}
	if errors.Is(err, ErrBadRequest) {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrTooManyRequests) {
		return http.StatusTooManyRequests
	}
	if errors.Is(err, ErrValidation) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

// PublicMessage returns a safe message to expose to API consumers.
// In production, internal details are hidden.
func PublicMessage(err error, env string) string {
	if err == nil {
		return ""
	}
	var ae *AppError
	if errors.As(err, &ae) && ae.Message != "" {
		return ae.Message
	}
	if env == "production" {
		return messages.MsgInternalError
	}
	return err.Error()
}
