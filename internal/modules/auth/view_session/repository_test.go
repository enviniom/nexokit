package view_session

import (
	"testing"

	"github.com/enviniom/nexokit/internal/platform/authctx"
)

func TestRepository_BuildSession(t *testing.T) {
	repo := NewRepository()
	companyID := uint(7)
	current := &authctx.User{
		PublicID:    "user-7",
		Name:        "Alice",
		Email:       "alice@example.com",
		RoleID:      2,
		Role:        "admin",
		RoleSlug:    "admin",
		CompanyID:   &companyID,
		IsActive:    true,
		Permissions: []string{"users.list", "roles.view"},
	}

	view, err := repo.BuildSession(current)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.PublicID != "user-7" || view.RoleSlug != "admin" || len(view.Permissions) != 2 {
		t.Fatalf("unexpected session view: %#v", view)
	}
}
