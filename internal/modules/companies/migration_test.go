package companies

import (
	"os"
	"strings"
	"testing"
)

func TestCompaniesMigrationDefinesCompanyTableAndUserCompanyReference(t *testing.T) {
	content, err := os.ReadFile("../../../migrations/20260519000000_companies.sql")
	if err != nil {
		t.Fatalf("expected companies migration to exist: %v", err)
	}
	sql := strings.ToLower(string(content))

	for _, required := range []string{
		"create table if not exists companies",
		"public_id char(26) not null unique",
		"slug varchar(120) not null unique",
		"status varchar(20) not null",
		"foreign key (company_id) references companies(id)",
		"create index if not exists idx_users_company_id",
		"drop table if exists companies",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("expected migration to contain %q, got:\n%s", required, string(content))
		}
	}
}
