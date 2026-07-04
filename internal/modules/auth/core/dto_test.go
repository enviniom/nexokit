package core

import (
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestLoginRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      LoginRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: LoginRequest{
				Email:    "alice@example.com",
				Password: "Secret1!",
			},
		},
		{
			name: "all required fields missing",
			req:  LoginRequest{},
			wantMsgs: map[string]string{
				"email":    messages.MsgRequired,
				"password": messages.MsgRequired,
			},
		},
		{
			name: "email invalid",
			req: LoginRequest{
				Email:    "not-an-email",
				Password: "Secret1!",
			},
			wantMsgs: map[string]string{
				"email": messages.MsgValidEmail,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestRefreshRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      RefreshRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req:  RefreshRequest{RefreshToken: "refresh-token"},
		},
		{
			name: "missing refresh token",
			req:  RefreshRequest{},
			wantMsgs: map[string]string{
				"refresh_token": messages.MsgRequired,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func assertValidation(t *testing.T, errs map[string][]string, wantMsgs map[string]string) {
	t.Helper()

	if len(wantMsgs) == 0 {
		if len(errs) > 0 {
			t.Fatalf("expected no validation errors, got %v", errs)
		}
		return
	}

	if len(errs) == 0 {
		t.Fatalf("expected validation errors, got none")
	}

	for field, wantMsg := range wantMsgs {
		fieldErrs, ok := errs[field]
		if !ok {
			t.Errorf("expected error for field %q, got none", field)
			continue
		}
		if len(fieldErrs) == 0 {
			t.Errorf("expected non-empty error list for field %q", field)
			continue
		}

		joined := joinErrors(fieldErrs)
		if !contains(joined, wantMsg) {
			t.Errorf("field %q: expected message containing %q, got %q", field, wantMsg, joined)
		}
	}
}

func joinErrors(errs []string) string {
	s := ""
	for i, e := range errs {
		if i > 0 {
			s += "; "
		}
		s += e
	}
	return s
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(haystack) > 0 && containsSubstring(haystack, needle))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
