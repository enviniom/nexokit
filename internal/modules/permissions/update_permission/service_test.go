package update_permission

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

type fakeRepo struct {
	item *core.Permission
	err  error
}

func (f fakeRepo) GetByPublicID(string) (*core.Permission, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}
func (f fakeRepo) Update(*core.Permission) error { return f.err }

func TestServiceReturnsSystemConflict(t *testing.T) {
	svc := NewService(fakeRepo{item: &core.Permission{IsSystem: true}})
	_, err := svc.Update("p1", core.UpdatePermissionRequest{Name: "n"})
	if !errors.Is(err, core.ErrSystemImmutable) {
		t.Fatalf("expected ErrSystemImmutable, got %v", err)
	}
}

func TestServiceMapsNotFound(t *testing.T) {
	svc := NewService(fakeRepo{err: gorm.ErrRecordNotFound})
	_, err := svc.Update("missing", core.UpdatePermissionRequest{Name: "n"})
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceUpdatesFields(t *testing.T) {
	p := &core.Permission{BaseModel: shared.BaseModel{PublicID: "p1"}, Name: "old"}
	svc := NewService(fakeRepo{item: p})
	r, err := svc.Update("p1", core.UpdatePermissionRequest{Name: "new", Description: "d", DisplayOrder: 3})
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "new" || r.DisplayOrder != 3 {
		t.Fatalf("unexpected update: %+v", r)
	}
}
