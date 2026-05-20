package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultitenancyGuideCoversTenantModelRepositoryScopeAndReviewChecklist(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "docs", "multitenancy.md"))
	if err != nil {
		t.Fatalf("reading multitenancy guide: %v", err)
	}

	text := string(content)
	for _, want := range []string{
		"## Tenant model fields",
		"CompanyID",
		"CompanySlug",
		"IsRootScope",
		"## Repository scope rules",
		"ApplyTenantScope",
		"Cross-tenant misses return 404",
		"## Review checklist for new modules",
		"[ ] Tenant-owned models include `company_id`",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("multitenancy guide missing %q", want)
		}
	}
}
