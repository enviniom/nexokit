package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestCodeConstants(t *testing.T) {
	cases := []struct {
		code   Code
		status int
	}{
		{CodeNotFound, http.StatusNotFound},
		{CodeBadRequest, http.StatusBadRequest},
		{CodeForbidden, http.StatusForbidden},
		{CodeConflict, http.StatusConflict},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeTooManyRequests, http.StatusTooManyRequests},
		{CodeValidation, http.StatusUnprocessableEntity},
		{CodeUnprocessable, http.StatusUnprocessableEntity},
		{CodeInternal, http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(string(tc.code), func(t *testing.T) {
			if tc.code == "" {
				t.Fatalf("code must not be empty")
			}
			if tc.code.Error() != string(tc.code) {
				t.Errorf("Code.Error() = %q, want %q", tc.code.Error(), string(tc.code))
			}
		})
	}
}

func TestHelpersSetCodeAndStatus(t *testing.T) {
	customCode := Code("custom_missing")
	internalErr := errors.New("db boom")

	cases := []struct {
		name       string
		err        *AppError
		wantStatus int
		wantCode   Code
		wantMsg    string
	}{
		{"NotFound", NotFound(customCode, "user not found", internalErr), http.StatusNotFound, customCode, "user not found"},
		{"BadRequest", BadRequest(customCode, "invalid", internalErr), http.StatusBadRequest, customCode, "invalid"},
		{"Forbidden", Forbidden(customCode, "denied", internalErr), http.StatusForbidden, customCode, "denied"},
		{"Conflict", Conflict(customCode, "taken", internalErr), http.StatusConflict, customCode, "taken"},
		{"Unauthorized", Unauthorized(customCode, "no auth", internalErr), http.StatusUnauthorized, customCode, "no auth"},
		{"TooManyRequests", TooManyRequests(customCode, "rate limited", internalErr), http.StatusTooManyRequests, customCode, "rate limited"},
		{"Validation", Validation(customCode, "bad input", internalErr), http.StatusUnprocessableEntity, customCode, "bad input"},
		{"Unprocessable", Unprocessable(customCode, "cannot process", internalErr), http.StatusUnprocessableEntity, customCode, "cannot process"},
		{"Internal", Internal(customCode, "boom", internalErr), http.StatusInternalServerError, customCode, "boom"},
		{"New", New(customCode, http.StatusPaymentRequired, "pay up", internalErr), http.StatusPaymentRequired, customCode, "pay up"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.HTTPStatus != tc.wantStatus {
				t.Errorf("HTTPStatus = %d, want %d", tc.err.HTTPStatus, tc.wantStatus)
			}
			if tc.err.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", tc.err.Code, tc.wantCode)
			}
			if tc.err.PublicMessage != tc.wantMsg {
				t.Errorf("PublicMessage = %q, want %q", tc.err.PublicMessage, tc.wantMsg)
			}
			if tc.err.Internal != internalErr {
				t.Errorf("Internal = %v, want %v", tc.err.Internal, internalErr)
			}
		})
	}
}

func TestInternalIsUnwrapSource(t *testing.T) {
	original := errors.New("original")
	wrapped := fmt.Errorf("layer: %w", original)
	ae := NotFound(CodeNotFound, "foo", wrapped)

	if !errors.Is(ae, original) {
		t.Error("expected errors.Is to reach original via Internal chain")
	}
	if !errors.Is(ae, wrapped) {
		t.Error("expected errors.Is to reach wrapped intermediate")
	}
}

func TestIsPointerMatch(t *testing.T) {
	ae := NotFound(CodeNotFound, "foo", nil)
	if !errors.Is(ae, ae) {
		t.Error("expected pointer identity match")
	}
}

func TestIsCodeEquality(t *testing.T) {
	ae := NotFound(CodeNotFound, "user not found", errors.New("db"))

	if !errors.Is(ae, CodeNotFound) {
		t.Error("expected errors.Is(ae, CodeNotFound) to match by code")
	}
	if !errors.Is(ae, ErrNotFound) {
		t.Error("expected errors.Is(ae, ErrNotFound) to match by code equality")
	}
}

func TestIsFallsThroughToInternal(t *testing.T) {
	inner := errors.New("inner")
	ae := NotFound(CodeNotFound, "foo", inner)

	if !errors.Is(ae, inner) {
		t.Error("expected errors.Is to fall through to Internal")
	}
}

func TestIsEmptyCodeDoesNotOvermatch(t *testing.T) {
	ae1 := &AppError{Code: Code("abc"), HTTPStatus: 400, PublicMessage: "x"}
	ae2 := &AppError{Code: Code("def"), HTTPStatus: 400, PublicMessage: "x"}
	if errors.Is(ae1, ae2) {
		t.Error("expected different codes not to match")
	}
}

func TestAsRecoversAppErrorFromWrappedChain(t *testing.T) {
	ae := NotFound(CodeNotFound, "foo", nil)
	wrapped := fmt.Errorf("layer: %w", ae)

	var recovered *AppError
	if !errors.As(wrapped, &recovered) {
		t.Fatal("expected errors.As to recover AppError")
	}
	if recovered.PublicMessage != "foo" {
		t.Errorf("PublicMessage = %q, want %q", recovered.PublicMessage, "foo")
	}
}

func TestWrapPreservesSentinelStatusAndCode(t *testing.T) {
	const roleMsg = "role has assigned users"
	wrapped := Wrap(ErrUnprocessable, roleMsg)

	if Status(wrapped) != http.StatusUnprocessableEntity {
		t.Errorf("Status = %d, want %d", Status(wrapped), http.StatusUnprocessableEntity)
	}
	if !errors.Is(wrapped, ErrUnprocessable) {
		t.Error("expected errors.Is(wrapped, ErrUnprocessable) to be true")
	}
	if !errors.Is(wrapped, CodeUnprocessable) {
		t.Error("expected errors.Is(wrapped, CodeUnprocessable) to be true")
	}
	if wrapped.PublicMessage != roleMsg {
		t.Errorf("PublicMessage = %q, want %q", wrapped.PublicMessage, roleMsg)
	}
}

func TestWrapPreservesNotFound(t *testing.T) {
	inner := errors.New("db boom")
	wrapped := Wrap(ErrNotFound, "user 123 not found", inner)

	if Status(wrapped) != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", Status(wrapped), http.StatusNotFound)
	}
	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("expected errors.Is(wrapped, ErrNotFound) to be true")
	}
	if !errors.Is(wrapped, inner) {
		t.Error("expected unwrap chain to reach inner error")
	}
}

func TestWrapUnknownDefaultsToInternal(t *testing.T) {
	wrapped := Wrap(errors.New("random failure"), "ctx")

	if Status(wrapped) != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", Status(wrapped), http.StatusInternalServerError)
	}
	if !errors.Is(wrapped, ErrInternal) {
		t.Error("expected errors.Is(wrapped, ErrInternal) to be true")
	}
	if !errors.Is(wrapped, CodeInternal) {
		t.Error("expected errors.Is(wrapped, CodeInternal) to be true")
	}
	if wrapped.Code != CodeInternal {
		t.Errorf("Code = %q, want %q", wrapped.Code, CodeInternal)
	}
}

func TestWrapWithCause(t *testing.T) {
	inner := errors.New("timeout")
	cause := errors.New("network unreachable")
	wrapped := Wrap(inner, "request failed", cause)

	if !errors.Is(wrapped, inner) {
		t.Error("expected inner error to be unwrappable")
	}
	if !errors.Is(wrapped, cause) {
		t.Error("expected cause to be unwrappable")
	}
}

func TestWrapWithMultipleCauses(t *testing.T) {
	inner := errors.New("timeout")
	cause1 := errors.New("network unreachable")
	cause2 := errors.New("dns failure")
	wrapped := Wrap(inner, "request failed", cause1, cause2)

	if !errors.Is(wrapped, inner) {
		t.Error("expected inner error to be unwrappable")
	}
	if !errors.Is(wrapped, cause1) {
		t.Error("expected cause1 to be unwrappable")
	}
	if !errors.Is(wrapped, cause2) {
		t.Error("expected cause2 to be unwrappable")
	}
}

func TestWrapNilDefaultsToInternal(t *testing.T) {
	wrapped := Wrap(nil, "ctx")
	if Status(wrapped) != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", Status(wrapped), http.StatusInternalServerError)
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
		{"too many requests", ErrTooManyRequests, http.StatusTooManyRequests},
		{"validation", ErrValidation, http.StatusUnprocessableEntity},
		{"unprocessable", ErrUnprocessable, http.StatusUnprocessableEntity},
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

func TestPublicMessage_ForAppErrorIgnoresMode(t *testing.T) {
	ae := NotFound(CodeNotFound, "user not found", errors.New("db"))
	for _, mode := range []string{"debug", "test", "release"} {
		if got := PublicMessage(ae, mode); got != "user not found" {
			t.Errorf("mode %q: PublicMessage = %q, want %q", mode, got, "user not found")
		}
	}
}

func TestPublicMessage_UnknownRedactsRegardlessOfMode(t *testing.T) {
	plain := errors.New("secret stack trace")
	for _, mode := range []string{"development", "test", "release"} {
		msg := PublicMessage(plain, mode)
		if msg != messages.MsgInternalError {
			t.Errorf("mode %q: expected generic message, got %s", mode, msg)
		}
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
	if ErrTooManyRequests.Error() != messages.MsgTooManyRequests {
		t.Errorf("unexpected too many requests sentinel message: %s", ErrTooManyRequests.Error())
	}
	if ErrUnprocessable.PublicMessage != "" {
		t.Errorf("expected ErrUnprocessable.PublicMessage to be empty, got %q", ErrUnprocessable.PublicMessage)
	}
	if Status(ErrUnprocessable) != http.StatusUnprocessableEntity {
		t.Errorf("expected ErrUnprocessable to map to 422, got %d", Status(ErrUnprocessable))
	}
}

func TestErrorString(t *testing.T) {
	ae := NotFound(CodeNotFound, "foo", errors.New("db"))
	if ae.Error() != "db" {
		t.Errorf("Error() = %q, want %q", ae.Error(), "db")
	}

	withoutInternal := NotFound(CodeNotFound, "foo", nil)
	if withoutInternal.Error() != "foo" {
		t.Errorf("Error() = %q, want %q", withoutInternal.Error(), "foo")
	}
}
