package update_permission

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct {
	item      *core.IAMPermission
	getErr    error
	updateErr error
}

func (f fakeRepo) GetPermissionByPublicID(string) (*core.IAMPermission, error) {
	return f.item, f.getErr
}
func (f fakeRepo) UpdatePermission(*core.IAMPermission) error { return f.updateErr }

func TestUpdateSuccess(t *testing.T) {
	repo := fakeRepo{item: &core.IAMPermission{BaseModel: shared.BaseModel{PublicID: "perm-1"}, Name: "Old", Description: "old", DisplayOrder: 1}}
	svc := NewService(repo)

	res, err := svc.Update("perm-1", core.UpdatePermissionRequest{Name: "New", Description: "new", DisplayOrder: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "New" || res.Description != "new" || res.DisplayOrder != 5 {
		t.Fatalf("unexpected response payload: %#v", res)
	}
}

func TestUpdateRejectsSystemPermission(t *testing.T) {
	svc := NewService(fakeRepo{item: &core.IAMPermission{IsSystem: true}})
	_, err := svc.Update("p1", core.UpdatePermissionRequest{Name: "x"})
	if !errors.Is(err, core.ErrSystemImmutable) {
		t.Fatalf("expected ErrSystemImmutable, got %v", err)
	}
}

func TestUpdateMapsNotFound(t *testing.T) {
	svc := NewService(fakeRepo{getErr: core.ErrNotFound})
	_, err := svc.Update("p1", core.UpdatePermissionRequest{Name: "x"})
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdatePropagatesConflict(t *testing.T) {
	svc := NewService(fakeRepo{item: &core.IAMPermission{}, updateErr: core.ErrConflict})
	_, err := svc.Update("p1", core.UpdatePermissionRequest{Name: "x"})
	if !errors.Is(err, core.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestUpdatePropagatesUnexpectedRepositoryError(t *testing.T) {
	repoErr := errors.New("repository unavailable")
	svc := NewService(fakeRepo{item: &core.IAMPermission{}, updateErr: repoErr})
	_, err := svc.Update("p1", core.UpdatePermissionRequest{Name: "x"})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}
