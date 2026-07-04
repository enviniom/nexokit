package core

import (
	"fmt"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestOnboardCompanyRequest_Validate(t *testing.T) {
	longName := "n"
	for i := 0; i < 100; i++ {
		longName += "a"
	}

	tests := []struct {
		name         string
		req          OnboardCompanyRequest
		wantFields   []string
		wantRequired bool
		wantMinLen   bool
		wantEmail    bool
	}{
		{
			name: "valid request",
			req: OnboardCompanyRequest{
				Name:          "Acme Inc",
				Slug:          "acme",
				AdminName:     "Jane Admin",
				AdminEmail:    "jane@acme.com",
				AdminPassword: "SecurePass123",
			},
		},
		{
			name: "all required fields missing",
			req:  OnboardCompanyRequest{},
			wantFields: []string{
				"name",
				"slug",
				"admin_name",
				"admin_email",
				"admin_password",
			},
			wantRequired: true,
		},
		{
			name: "name too short",
			req: OnboardCompanyRequest{
				Name:          "A",
				Slug:          "acme",
				AdminName:     "Jane Admin",
				AdminEmail:    "jane@acme.com",
				AdminPassword: "SecurePass123",
			},
			wantFields: []string{"name"},
			wantMinLen: true,
		},
		{
			name: "slug too short",
			req: OnboardCompanyRequest{
				Name:          "Acme Inc",
				Slug:          "a",
				AdminName:     "Jane Admin",
				AdminEmail:    "jane@acme.com",
				AdminPassword: "SecurePass123",
			},
			wantFields: []string{"slug"},
			wantMinLen: true,
		},
		{
			name: "admin name too short",
			req: OnboardCompanyRequest{
				Name:          "Acme Inc",
				Slug:          "acme",
				AdminName:     "J",
				AdminEmail:    "jane@acme.com",
				AdminPassword: "SecurePass123",
			},
			wantFields: []string{"admin_name"},
			wantMinLen: true,
		},
		{
			name: "admin email invalid",
			req: OnboardCompanyRequest{
				Name:          "Acme Inc",
				Slug:          "acme",
				AdminName:     "Jane Admin",
				AdminEmail:    "not-an-email",
				AdminPassword: "SecurePass123",
			},
			wantFields: []string{"admin_email"},
			wantEmail:  true,
		},
		{
			name: "admin password too short",
			req: OnboardCompanyRequest{
				Name:          "Acme Inc",
				Slug:          "acme",
				AdminName:     "Jane Admin",
				AdminEmail:    "jane@acme.com",
				AdminPassword: "short",
			},
			wantFields: []string{"admin_password"},
			wantMinLen: true,
		},
		{
			name: "multiple validation failures",
			req: OnboardCompanyRequest{
				Name:          "",
				Slug:          "a",
				AdminName:     "",
				AdminEmail:    "bad-email",
				AdminPassword: "tiny",
			},
			wantFields: []string{"name", "slug", "admin_name", "admin_email", "admin_password"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()

			if len(tt.wantFields) == 0 {
				if len(errs) > 0 {
					t.Fatalf("expected no validation errors, got %v", errs)
				}
				return
			}

			if len(errs) == 0 {
				t.Fatalf("expected validation errors for fields %v, got none", tt.wantFields)
			}

			for _, field := range tt.wantFields {
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

				if tt.wantRequired {
					if !contains(joined, messages.MsgRequired) {
						t.Errorf("field %q: expected required error %q, got %q", field, messages.MsgRequired, joined)
					}
				}

				if tt.wantMinLen {
					expected := fmt.Sprintf(messages.MsgMinLength, 2)
					if field == "admin_password" {
						expected = fmt.Sprintf(messages.MsgMinLength, 8)
					}
					if !contains(joined, expected) {
						t.Errorf("field %q: expected min-length error %q, got %q", field, expected, joined)
					}
				}

				if tt.wantEmail {
					if !contains(joined, messages.MsgValidEmail) {
						t.Errorf("field %q: expected email error %q, got %q", field, messages.MsgValidEmail, joined)
					}
				}
			}
		})
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
