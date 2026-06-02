package queries

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestExistsRoleBySlug(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	items := []core.IAMRole{
		{BaseModel: shared.BaseModel{PublicID: "role-root"}, Name: "Root Manager", Slug: "manager", CompanyID: nil},
		{BaseModel: shared.BaseModel{PublicID: "role-root-director"}, Name: "Root Director", Slug: "director", CompanyID: nil},
		{BaseModel: shared.BaseModel{PublicID: "role-acme-1"}, Name: "Acme Manager 1", Slug: "manager", CompanyID: uintPtr(10)},
		{BaseModel: shared.BaseModel{PublicID: "role-acme-2"}, Name: "Acme Manager 2", Slug: "manager", CompanyID: uintPtr(10)},
	}
	for i := range items {
		if err := db.Create(&items[i]).Error; err != nil {
			t.Fatalf("seed role: %v", err)
		}
	}

	tests := []struct {
		name          string
		tenant        tenant.TenantContext
		value         string
		excludeRoleID uint
		expected      bool
	}{
		{name: "exists without exclusion", tenant: tenant.NewRoot(), value: "manager", expected: true},
		{name: "false when only matching record is excluded", tenant: tenant.NewRoot(), value: "director", excludeRoleID: items[1].ID, expected: false},
		{name: "true when another matching record exists after exclusion", tenant: tenant.NewScoped(10, "acme"), value: "manager", excludeRoleID: items[2].ID, expected: true},
		{name: "scoped tenant sees own role", tenant: tenant.NewScoped(10, "acme"), value: "manager", expected: true},
		{name: "tenant scoping remains correct with exclusion", tenant: tenant.NewScoped(10, "acme"), value: "director", expected: false},
		{name: "scoped tenant does not see other tenant", tenant: tenant.NewScoped(20, "globex"), value: "manager", expected: false},
		{name: "missing slug returns false", tenant: tenant.NewRoot(), value: "missing", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := ExistsRoleBySlug(db, tt.tenant, tt.value, tt.excludeRoleID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, exists)
			}
		})
	}
}
