package queries

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestExistsRoleByName(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&core.IAMRole{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	items := []core.IAMRole{
		{BaseModel: shared.BaseModel{PublicID: "role-root"}, Name: "Manager", Slug: "manager", CompanyID: nil},
		{BaseModel: shared.BaseModel{PublicID: "role-root-director"}, Name: "Director", Slug: "director", CompanyID: nil},
		{BaseModel: shared.BaseModel{PublicID: "role-acme-1"}, Name: "Manager", Slug: "manager-acme-1", CompanyID: uintPtr(10)},
		{BaseModel: shared.BaseModel{PublicID: "role-acme-2"}, Name: "Manager", Slug: "manager-acme-2", CompanyID: uintPtr(10)},
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
		{name: "exists without exclusion", tenant: tenant.NewRoot(), value: "Manager", expected: true},
		{name: "false when only matching record is excluded", tenant: tenant.NewRoot(), value: "Director", excludeRoleID: items[1].ID, expected: false},
		{name: "true when another matching record exists after exclusion", tenant: tenant.NewScoped(10, "acme"), value: "Manager", excludeRoleID: items[2].ID, expected: true},
		{name: "scoped tenant sees own role", tenant: tenant.NewScoped(10, "acme"), value: "Manager", expected: true},
		{name: "tenant scoping remains correct with exclusion", tenant: tenant.NewScoped(10, "acme"), value: "Director", expected: false},
		{name: "scoped tenant does not see other tenant", tenant: tenant.NewScoped(20, "globex"), value: "Manager", expected: false},
		{name: "missing name returns false", tenant: tenant.NewRoot(), value: "Missing", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := ExistsRoleByName(db, tt.tenant, tt.value, tt.excludeRoleID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, exists)
			}
		})
	}
}

func uintPtr(v uint) *uint { return &v }
