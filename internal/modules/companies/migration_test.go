package companies

import (
	"os"
	"strings"
	"testing"
)

func TestCompaniesMigrationDefinesCompanyTableAndUserCompanyReference(t *testing.T) {
	content, err := os.ReadFile("../../../migrations/20260101000000_init.sql")
	if err != nil {
		t.Fatalf("expected consolidated migration to exist: %v", err)
	}
	sql := strings.ToLower(string(content))

	for _, required := range []string{
		"create table if not exists companies",
		"public_id char(26) not null unique",
		"slug varchar(120) not null unique",
		"status varchar(20) not null",
		"create table if not exists company_domains",
		"domain varchar(255) not null unique",
		"kind varchar(40) not null",
		"redirect_to_primary boolean not null default false",
		"create index if not exists idx_company_domains_company_id",
		"references companies(id)",
		"create index if not exists idx_users_company_id",
		"drop table if exists company_domains",
		"drop table if exists companies",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("expected migration to contain %q, got:\n%s", required, string(content))
		}
	}

	for _, forbidden := range []string{
		"domain varchar(255),",
		"subdomain varchar(120)",
		"idx_companies_domain",
		"idx_companies_subdomain",
	} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("expected migration not to contain %q, got:\n%s", forbidden, string(content))
		}
	}
}
