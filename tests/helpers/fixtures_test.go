package helpers

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
)

func TestFixtures_SeedRelationshipAwareData(t *testing.T) {
	t.Parallel()

	db := NewSQLiteDB(t, &iamcore.IAMRole{}, &companies.Company{}, &iamcore.IAMUser{}, &iamcore.IAMPermission{})

	company := SeedCompany(t, db, "acme")
	role := SeedRole(t, db, iamcore.AdminRoleSlug)
	user := SeedUserWithRole(t, db, "user-acme", role, &company)
	perm := SeedPermission(t, db, "users", platformPerms.ActionList)

	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "user belongs to seeded company",
			check: func(t *testing.T) {
				if user.CompanyID == nil || *user.CompanyID != company.ID {
					t.Fatalf("expected user to belong to company %d, got %+v", company.ID, user.CompanyID)
				}
			},
		},
		{
			name: "user keeps seeded role",
			check: func(t *testing.T) {
				if user.RoleID != role.ID {
					t.Fatalf("expected user role %d, got %d", role.ID, user.RoleID)
				}
			},
		},
		{
			name: "permission slug is deterministic",
			check: func(t *testing.T) {
				if perm.Slug != "users.list" {
					t.Fatalf("expected deterministic permission slug users.list, got %s", perm.Slug)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.check(t)
		})
	}
}
