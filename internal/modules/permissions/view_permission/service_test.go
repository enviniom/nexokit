package view_permission

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

func TestServiceMapsNotFoundToCoreError(t *testing.T) {
	svc := NewService(fakeRepo{err: gorm.ErrRecordNotFound})
	_, err := svc.GetByPublicID("missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected core.ErrNotFound, got %v", err)
	}
}

func TestServiceReturnsPermission(t *testing.T) {
	svc := NewService(fakeRepo{item: &core.Permission{BaseModel: shared.BaseModel{PublicID: "p1"}, Slug: "users.list"}})
	resp, err := svc.GetByPublicID("p1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.PublicID != "p1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
