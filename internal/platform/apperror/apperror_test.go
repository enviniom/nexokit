package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestWrap(t *testing.T) {
	inner := errors.New("database connection lost")
	wrapped := Wrap(inner, "failed to load user")

	if wrapped.Err != inner {
		t.Error("expected Err to be the inner error")
	}
	if wrapped.Message != "failed to load user" {
		t.Errorf("unexpected message: %s", wrapped.Message)
	}
	if !errors.Is(wrapped, inner) {
		t.Error("expected wrapped error to be identifiable with errors.Is")
	}
}

func TestWrapWithCause(t *testing.T) {
	inner := errors.New("timeout")
	cause := errors.New("network unreachable")
	wrapped := Wrap(inner, "request failed", cause)

	if !errors.Is(wrapped, cause) {
		t.Error("expected cause to be unwrappable")
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect int
	}{
		{"nil", nil, http.StatusOK},
		{"not found", ErrNotFound, http.StatusNotFound},
		{"forbidden", ErrForbidden, http.StatusForbidden},
		{"unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"conflict", ErrConflict, http.StatusConflict},
		{"bad request", ErrBadRequest, http.StatusBadRequest},
		{"validation", ErrValidation, http.StatusUnprocessableEntity},
		{"internal", ErrInternal, http.StatusInternalServerError},
		{"unknown", errors.New("something else"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Status(tt.err); got != tt.expect {
				t.Errorf("Status(%v) = %d; want %d", tt.err, got, tt.expect)
			}
		})
	}
}

func TestStatus_WrappedError(t *testing.T) {
	wrapped := Wrap(ErrNotFound, "user lookup failed")
	if Status(wrapped) != http.StatusNotFound {
		t.Errorf("expected %d, got %d", http.StatusNotFound, Status(wrapped))
	}
}

func TestPublicMessage(t *testing.T) {
	wrapped := Wrap(errors.New("db timeout"), "failed to fetch data")
	msg := PublicMessage(wrapped, "development")
	if msg != "failed to fetch data" {
		t.Errorf("expected 'failed to fetch data', got %s", msg)
	}
}

func TestPublicMessage_Production(t *testing.T) {
	plain := errors.New("secret stack trace")
	msg := PublicMessage(plain, "production")
	if msg != messages.MsgInternalError {
		t.Errorf("expected generic message in production, got %s", msg)
	}
}

func TestPublicMessage_Nil(t *testing.T) {
	if PublicMessage(nil, "development") != "" {
		t.Error("expected empty message for nil error")
	}
}

func TestSentinels(t *testing.T) {
	if ErrNotFound.Error() != messages.MsgNotFound {
		t.Errorf("unexpected sentinel message: %s", ErrNotFound.Error())
	}
	if ErrForbidden.Error() != messages.MsgForbidden {
		t.Errorf("unexpected sentinel message: %s", ErrForbidden.Error())
	}
	if !errors.Is(ErrValidation, ErrValidation) {
		t.Error("expected ErrValidation to match with errors.Is")
	}
	if ErrValidation.Error() != messages.MsgValidationError {
		t.Errorf("unexpected validation sentinel message: %s", ErrValidation.Error())
	}
}
