package core

import (
	"fmt"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestCreateCompanyRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      CreateCompanyRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: CreateCompanyRequest{
				Name:   "Acme Inc",
				Slug:   "acme",
				Status: CompanyStatusActive,
			},
		},
		{
			name: "all required fields missing",
			req:  CreateCompanyRequest{},
			wantMsgs: map[string]string{
				"name": messages.MsgRequired,
				"slug": messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: CreateCompanyRequest{
				Name:   "A",
				Slug:   "acme",
				Status: CompanyStatusActive,
			},
			wantMsgs: map[string]string{
				"name": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "slug too short",
			req: CreateCompanyRequest{
				Name:   "Acme Inc",
				Slug:   "a",
				Status: CompanyStatusActive,
			},
			wantMsgs: map[string]string{
				"slug": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "invalid status",
			req: CreateCompanyRequest{
				Name:   "Acme Inc",
				Slug:   "acme",
				Status: "invalid",
			},
			wantMsgs: map[string]string{
				"status": messages.MsgInvalidFormat,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestUpdateCompanyRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      UpdateCompanyRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: UpdateCompanyRequest{
				Name:   "Acme Inc",
				Slug:   "acme",
				Status: CompanyStatusActive,
			},
		},
		{
			name: "all required fields missing",
			req:  UpdateCompanyRequest{},
			wantMsgs: map[string]string{
				"name": messages.MsgRequired,
				"slug": messages.MsgRequired,
			},
		},
		{
			name: "name too short",
			req: UpdateCompanyRequest{
				Name:   "A",
				Slug:   "acme",
				Status: CompanyStatusActive,
			},
			wantMsgs: map[string]string{
				"name": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "slug too short",
			req: UpdateCompanyRequest{
				Name:   "Acme Inc",
				Slug:   "a",
				Status: CompanyStatusActive,
			},
			wantMsgs: map[string]string{
				"slug": fmt.Sprintf(messages.MsgMinLength, 2),
			},
		},
		{
			name: "invalid status",
			req: UpdateCompanyRequest{
				Name:   "Acme Inc",
				Slug:   "acme",
				Status: "invalid",
			},
			wantMsgs: map[string]string{
				"status": messages.MsgInvalidFormat,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestCreateCompanyDomainRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      CreateCompanyDomainRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid primary request",
			req: CreateCompanyDomainRequest{
				Domain: "acme.com",
				Kind:   CompanyDomainKindPrimary,
				Status: CompanyDomainStatusActive,
			},
		},
		{
			name: "valid alias request",
			req: CreateCompanyDomainRequest{
				Domain:            "www.acme.com",
				Kind:              CompanyDomainKindAlias,
				Status:            CompanyDomainStatusActive,
				RedirectToPrimary: true,
			},
		},
		{
			name: "missing fields",
			req:  CreateCompanyDomainRequest{},
			wantMsgs: map[string]string{
				"domain": messages.MsgRequired,
				"kind":   messages.MsgInvalidFormat,
				"status": messages.MsgInvalidFormat,
			},
		},
		{
			name: "domain too short",
			req: CreateCompanyDomainRequest{
				Domain: "a",
				Kind:   CompanyDomainKindPrimary,
				Status: CompanyDomainStatusActive,
			},
			wantMsgs: map[string]string{
				"domain": fmt.Sprintf(messages.MsgMinLength, 3),
			},
		},
		{
			name: "invalid kind",
			req: CreateCompanyDomainRequest{
				Domain: "acme.com",
				Kind:   "invalid",
				Status: CompanyDomainStatusActive,
			},
			wantMsgs: map[string]string{
				"kind": messages.MsgInvalidFormat,
			},
		},
		{
			name: "invalid status",
			req: CreateCompanyDomainRequest{
				Domain: "acme.com",
				Kind:   CompanyDomainKindPrimary,
				Status: "invalid",
			},
			wantMsgs: map[string]string{
				"status": messages.MsgInvalidFormat,
			},
		},
		{
			name: "redirect on primary kind",
			req: CreateCompanyDomainRequest{
				Domain:            "acme.com",
				Kind:              CompanyDomainKindPrimary,
				Status:            CompanyDomainStatusActive,
				RedirectToPrimary: true,
			},
			wantMsgs: map[string]string{
				"redirect_to_primary": messages.MsgInvalidFormat,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertValidation(t, tt.req.Validate(), tt.wantMsgs)
		})
	}
}

func TestUpdateCompanyDomainRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		req      UpdateCompanyDomainRequest
		wantMsgs map[string]string
	}{
		{
			name: "valid request",
			req: UpdateCompanyDomainRequest{
				Domain: "www.acme.com",
				Kind:   CompanyDomainKindAlias,
				Status: CompanyDomainStatusActive,
			},
		},
		{
			name: "missing fields",
			req:  UpdateCompanyDomainRequest{},
			wantMsgs: map[string]string{
				"domain": messages.MsgRequired,
				"kind":   messages.MsgInvalidFormat,
				"status": messages.MsgInvalidFormat,
			},
		},
		{
			name: "domain too short",
			req: UpdateCompanyDomainRequest{
				Domain: "a",
				Kind:   CompanyDomainKindPrimary,
				Status: CompanyDomainStatusActive,
			},
			wantMsgs: map[string]string{
				"domain": fmt.Sprintf(messages.MsgMinLength, 3),
			},
		},
		{
			name: "redirect on primary kind",
			req: UpdateCompanyDomainRequest{
				Domain:            "acme.com",
				Kind:              CompanyDomainKindPrimary,
				Status:            CompanyDomainStatusActive,
				RedirectToPrimary: true,
			},
			wantMsgs: map[string]string{
				"redirect_to_primary": messages.MsgInvalidFormat,
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
