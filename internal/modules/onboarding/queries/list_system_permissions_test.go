package queries

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestListSystemPermissions(t *testing.T) {
	db := newTestDB(t)

	seed := []core.OnboardingPermission{
		{BaseModel: shared.BaseModel{PublicID: "per_01"}, Slug: "users.view", Name: "View users", Module: "users", Action: "view", IsSystem: true},
		{BaseModel: shared.BaseModel{PublicID: "per_02"}, Slug: "users.create", Name: "Create users", Module: "users", Action: "create", IsSystem: false},
		{BaseModel: shared.BaseModel{PublicID: "per_03"}, Slug: "roles.view", Name: "View roles", Module: "roles", Action: "view", IsSystem: true},
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed permissions: %v", err)
	}

	permissions, err := ListSystemPermissions(db)
	if err != nil {
		t.Fatalf("query permissions: %v", err)
	}
	if len(permissions) != 2 {
		t.Fatalf("expected 2 system permissions, got %d", len(permissions))
	}
}
