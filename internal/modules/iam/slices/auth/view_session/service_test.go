package view_session

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/authctx"
)

type fakeServiceRepository struct{}

func (fakeServiceRepository) BuildSession(current *authctx.User) (*user.SessionView, error) {
	return &user.SessionView{RoleSlug: current.RoleSlug, Permissions: current.Permissions}, nil
}

func TestService_View(t *testing.T) {
	svc := NewService(fakeServiceRepository{})
	current := &authctx.User{RoleSlug: "admin", Permissions: []string{"users.list"}}

	view, err := svc.View(current)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.RoleSlug != "admin" || len(view.Permissions) != 1 {
		t.Fatalf("unexpected view: %#v", view)
	}
}
