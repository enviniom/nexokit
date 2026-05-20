package tenant

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type tenantScopedRecord struct {
	ID        uint
	CompanyID uint
}

func TestApplyTenantScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	cases := []struct {
		name             string
		tenant           TenantContext
		wantCompanyWhere bool
		wantVars         []any
	}{
		{
			name:             "non-root tenant filters by company_id",
			tenant:           NewScoped(3, "acme"),
			wantCompanyWhere: true,
			wantVars:         []any{uint(3)},
		},
		{
			name:             "root tenant remains unfiltered",
			tenant:           NewRoot(),
			wantCompanyWhere: false,
			wantVars:         nil,
		},
		{
			name:             "zero-company non-root is still scoped to company zero",
			tenant:           NewScoped(0, ""),
			wantCompanyWhere: true,
			wantVars:         []any{uint(0)},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			stmt := ApplyTenantScope(db.Model(&tenantScopedRecord{}), tt.tenant).Find(&[]tenantScopedRecord{}).Statement
			sql := stmt.SQL.String()

			hasCompanyWhere := strings.Contains(sql, "company_id")
			if hasCompanyWhere != tt.wantCompanyWhere {
				t.Fatalf("company_id filter presence = %v, want %v; sql=%s", hasCompanyWhere, tt.wantCompanyWhere, sql)
			}
			if len(stmt.Vars) != len(tt.wantVars) {
				t.Fatalf("vars = %#v, want %#v", stmt.Vars, tt.wantVars)
			}
			for i := range tt.wantVars {
				if stmt.Vars[i] != tt.wantVars[i] {
					t.Fatalf("vars = %#v, want %#v", stmt.Vars, tt.wantVars)
				}
			}
		})
	}
}

func TestGinTenantContext(t *testing.T) {
	cases := []struct {
		name       string
		stored     TenantContext
		wantTenant TenantContext
	}{
		{name: "scoped tenant round-trips", stored: NewScoped(9, "acme"), wantTenant: NewScoped(9, "acme")},
		{name: "root tenant round-trips", stored: NewRoot(), wantTenant: NewRoot()},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(nil)
			SetGin(ctx, tt.stored)

			got, ok := FromGin(ctx)
			if !ok {
				t.Fatal("expected tenant context to be present")
			}
			if got != tt.wantTenant {
				t.Fatalf("tenant = %#v, want %#v", got, tt.wantTenant)
			}
		})
	}
}
