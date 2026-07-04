package apperror

import (
	"errors"
	"net/http"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

// Code identifies the error category. It is the primary identity used by
// errors.Is when matching AppError values, and it is also an error itself so
// callers may write errors.Is(err, apperror.CodeNotFound).
type Code string

// Error implements the error interface so a Code can be used as an errors.Is
// target.
func (c Code) Error() string { return string(c) }

// AppError represents an application-level error with a client-safe message,
// an HTTP status, and an optional internal error chain for logging.
type AppError struct {
	Code          Code
	HTTPStatus    int
	PublicMessage string
	Internal      error
}

// Error returns the internal error text when available, otherwise the public
// message. This keeps sentinels usable as regular errors while keeping the
// client-visible text separate.
func (ae *AppError) Error() string {
	if ae.Internal != nil {
		return ae.Internal.Error()
	}
	return ae.PublicMessage
}

// Unwrap returns the internal error chain.
func (ae *AppError) Unwrap() error {
	return ae.Internal
}

// Is implements error matching for errors.Is.
// Matching order: pointer identity, non-empty Code equality, then the Internal
// chain. Code equality preserves existing sentinel compatibility while the
// Internal fall-through supports wrapped domain errors.
func (ae *AppError) Is(target error) bool {
	if ae == target {
		return true
	}
	if ae.Code != "" {
		switch t := target.(type) {
		case Code:
			if ae.Code == t {
				return true
			}
		case *AppError:
			if t.Code != "" && ae.Code == t.Code {
				return true
			}
		}
	}
	if ae.Internal != nil {
		return errors.Is(ae.Internal, target)
	}
	return false
}

// HTTP-category codes owned by the platform. Modules MUST declare their own
// business codes; these are generic fallbacks.
const (
	CodeNotFound        Code = "not_found"
	CodeBadRequest      Code = "bad_request"
	CodeForbidden       Code = "forbidden"
	CodeConflict        Code = "conflict"
	CodeUnauthorized    Code = "unauthorized"
	CodeTooManyRequests Code = "too_many_requests"
	CodeValidation      Code = "validation"
	CodeUnprocessable   Code = "unprocessable"
	CodeInternal        Code = "internal"
)

// Sentinel errors for common application states.
var (
	ErrNotFound        = &AppError{Code: CodeNotFound, HTTPStatus: http.StatusNotFound, PublicMessage: messages.MsgNotFound}
	ErrForbidden       = &AppError{Code: CodeForbidden, HTTPStatus: http.StatusForbidden, PublicMessage: messages.MsgForbidden}
	ErrUnauthorized    = &AppError{Code: CodeUnauthorized, HTTPStatus: http.StatusUnauthorized, PublicMessage: messages.MsgUnauthorized}
	ErrConflict        = &AppError{Code: CodeConflict, HTTPStatus: http.StatusConflict, PublicMessage: messages.MsgConflict}
	ErrBadRequest      = &AppError{Code: CodeBadRequest, HTTPStatus: http.StatusBadRequest, PublicMessage: messages.MsgBadRequest}
	ErrTooManyRequests = &AppError{Code: CodeTooManyRequests, HTTPStatus: http.StatusTooManyRequests, PublicMessage: messages.MsgTooManyRequests}
	ErrValidation      = &AppError{Code: CodeValidation, HTTPStatus: http.StatusUnprocessableEntity, PublicMessage: messages.MsgValidationError}
	ErrUnprocessable   = &AppError{Code: CodeUnprocessable, HTTPStatus: http.StatusUnprocessableEntity, PublicMessage: ""}
	ErrInternal        = &AppError{Code: CodeInternal, HTTPStatus: http.StatusInternalServerError, PublicMessage: messages.MsgInternalError}
)

// sentinels is the ordered list of known HTTP-category sentinels used by Wrap
// to preserve status/code for plain errors that match a known sentinel.
var sentinels = []*AppError{
	ErrNotFound,
	ErrForbidden,
	ErrUnauthorized,
	ErrConflict,
	ErrBadRequest,
	ErrTooManyRequests,
	ErrValidation,
	ErrUnprocessable,
	ErrInternal,
}

// New creates a low-level AppError. Prefer the named HTTP helpers.
func New(code Code, status int, publicMsg string, internal error) *AppError {
	return &AppError{
		Code:          code,
		HTTPStatus:    status,
		PublicMessage: publicMsg,
		Internal:      internal,
	}
}

// NotFound creates a 404 AppError.
func NotFound(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusNotFound, publicMsg, internal)
}

// BadRequest creates a 400 AppError.
func BadRequest(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusBadRequest, publicMsg, internal)
}

// Forbidden creates a 403 AppError.
func Forbidden(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusForbidden, publicMsg, internal)
}

// Conflict creates a 409 AppError.
func Conflict(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusConflict, publicMsg, internal)
}

// Unauthorized creates a 401 AppError.
func Unauthorized(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusUnauthorized, publicMsg, internal)
}

// TooManyRequests creates a 429 AppError.
func TooManyRequests(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusTooManyRequests, publicMsg, internal)
}

// Validation creates a 422 AppError for validation-style outcomes.
// This does not replace DTO field-level validation responses.
func Validation(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusUnprocessableEntity, publicMsg, internal)
}

// Unprocessable creates a 422 AppError.
func Unprocessable(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusUnprocessableEntity, publicMsg, internal)
}

// Internal creates a 500 AppError.
func Internal(code Code, publicMsg string, internal error) *AppError {
	return New(code, http.StatusInternalServerError, publicMsg, internal)
}

// Wrap creates a new AppError wrapping an existing error. If err is an
// *AppError or matches a known sentinel via errors.Is, the returned AppError
// inherits its Code and HTTPStatus; otherwise it defaults to CodeInternal/500.
// The passed message becomes the PublicMessage. Optional cause arguments are
// appended to the unwrap chain after err.
func Wrap(err error, message string, cause ...error) *AppError {
	if err == nil {
		return Internal(CodeInternal, message, nil)
	}

	var code Code
	var status int

	var ae *AppError
	if errors.As(err, &ae) {
		code = ae.Code
		status = ae.HTTPStatus
	} else {
		matched := false
		for _, sentinel := range sentinels {
			if errors.Is(err, sentinel) {
				code = sentinel.Code
				status = sentinel.HTTPStatus
				matched = true
				break
			}
		}
		if !matched {
			code = CodeInternal
			status = http.StatusInternalServerError
		}
	}

	return &AppError{
		Code:          code,
		HTTPStatus:    status,
		PublicMessage: message,
		Internal:      chainErrors(err, cause...),
	}
}

// chainErrors builds a linear unwrap chain starting with err followed by each
// cause.
func chainErrors(err error, cause ...error) error {
	if len(cause) == 0 {
		return err
	}
	return errorChain(append([]error{err}, cause...))
}

// errorChain is a private error type whose Unwrap method returns the next
// element in the slice. It implements Is so the first element can still be
// matched directly.
type errorChain []error

func (ec errorChain) Error() string {
	if len(ec) == 0 {
		return ""
	}
	return ec[0].Error()
}

func (ec errorChain) Unwrap() error {
	if len(ec) <= 1 {
		return nil
	}
	return errorChain(ec[1:])
}

func (ec errorChain) Is(target error) bool {
	for _, e := range ec {
		if errors.Is(e, target) {
			return true
		}
	}
	return false
}

// Status returns the appropriate HTTP status code for an error.
func Status(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.HTTPStatus
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
	if errors.Is(err, ErrUnprocessable) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

// PublicMessage returns a safe message to expose to API consumers.
// *AppError values always expose their PublicMessage. Any other error is
// redacted to messages.MsgInternalError regardless of mode. The mode parameter
// is kept for API compatibility but is no longer used for redaction decisions;
// redaction is encoded in the AppError contract and debug is owned by the
// response layer.
func PublicMessage(err error, mode string) string {
	_ = mode
	if err == nil {
		return ""
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.PublicMessage
	}
	return messages.MsgInternalError
}
