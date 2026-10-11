package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultitenancyGuideCoversTenantModelRepositoryScopeAndReviewChecklist(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "references", "infrastructure", "multitenancy.md"))
	if err != nil {
		t.Fatalf("reading multitenancy guide: %v", err)
	}

	text := string(content)
	for _, tt := range []struct {
		name string
		want string
	}{
		{"tenant context fields", "`TenantContext` has exactly `CompanyID uint`, `CompanySlug string`, `IsRootScope bool`"},
		{"repository scope", "Apply `tenant.ApplyTenantScope` before every tenant-owned read/write"},
		{"company filter", "it adds `WHERE company_id = ?` unless `IsRootScope` is true"},
		{"cross-tenant not found", "Cross-tenant misses return 404 through the module not-found outcome"},
		{"non-disclosure", "without disclosing whether the foreign-company record exists"},
		{"tenant ownership checklist", "[ ] Tenant-owned models include `company_id` with appropriate migration indexes"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(text, tt.want) {
				t.Errorf("multitenancy guide missing %q", tt.want)
			}
		})
	}
}
