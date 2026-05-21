package helpers

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/token"
	"gorm.io/gorm"
)

func TestAuthHelpers(t *testing.T) {
	t.Parallel()

	type DBState struct {
		db *gorm.DB
	}

	tests := []struct {
		name     string
		publicID string
		roleSlug string
		exercise func(t *testing.T, db *DBState, user users.User)
	}{
		{
			name:     "SeedUser creates persisted user",
			publicID: "u-auth-1",
			roleSlug: roles.AdminRoleSlug,
			exercise: func(t *testing.T, _ *DBState, user users.User) {
				if user.ID == 0 {
					t.Fatalf("expected persisted user id")
				}
				if user.CompanyID == nil {
					t.Fatalf("expected company assignment")
				}
			},
		},
		{
			name:     "CreateTestToken issues parseable access token",
			publicID: "u-auth-2",
			roleSlug: roles.UserRoleSlug,
			exercise: func(t *testing.T, dbState *DBState, user users.User) {
				tok := CreateTestToken(t, dbState.db, user)
				manager := token.NewManager("nexokit-test-secret", time.Hour)
				claims, err := manager.ParseAccess(tok)
				if err != nil {
					t.Fatalf("parse access token: %v", err)
				}
				if claims.Sub != user.PublicID {
					t.Fatalf("expected sub %s, got %s", user.PublicID, claims.Sub)
				}
			},
		},
		{
			name:     "AuthenticatedRequest adds bearer token",
			publicID: "u-auth-3",
			roleSlug: roles.AdminRoleSlug,
			exercise: func(t *testing.T, dbState *DBState, user users.User) {
				role := roles.Role{}
				if err := dbState.db.First(&role, user.RoleID).Error; err != nil {
					t.Fatalf("load user role: %v", err)
				}
				req := AuthenticatedRequest(t, "GET", "/users", nil, Actor{User: user, Role: role})
				auth := req.Header.Get("Authorization")
				if auth == "" {
					t.Fatalf("expected authorization header")
				}

				manager := token.NewManager("nexokit-test-secret", time.Hour)
				parsed, err := manager.ParseAccess(auth[len("Bearer "):])
				if err != nil {
					t.Fatalf("parse request token: %v", err)
				}
				if parsed.Sub != user.PublicID {
					t.Fatalf("expected token subject %s, got %s", user.PublicID, parsed.Sub)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := NewSQLiteDB(t, &roles.Role{}, &companies.Company{}, &users.User{})
			company := SeedCompany(t, db, "acme")
			user := SeedUser(t, db, UserOptions{PublicID: tt.publicID, Email: tt.publicID + "@example.com", Name: tt.publicID, RoleSlug: tt.roleSlug, CompanyID: &company.ID})
			state := &DBState{db: db}
			tt.exercise(t, state, user)
			if user.CompanyID == nil || *user.CompanyID != company.ID {
				t.Fatalf("expected company assignment %d, got %+v", company.ID, user.CompanyID)
			}
		})
	}
}
