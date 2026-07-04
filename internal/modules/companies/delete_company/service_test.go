package delete_company

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
)

type fakeDeleteRepo struct {
	deleted string
	err     error
}

func (f *fakeDeleteRepo) GetByPublicID(string) (*core.Company, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &core.Company{}, nil
}
func (f *fakeDeleteRepo) Delete(id string) error { f.deleted = id; return nil }

func TestService_Delete(t *testing.T) {
	repo := &fakeDeleteRepo{}
	svc := NewService(repo)
	if err := svc.Delete("cmp_01"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deleted != "cmp_01" {
		t.Fatalf("expected delete call with cmp_01, got %s", repo.deleted)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := &fakeDeleteRepo{err: core.ErrCompanyNotFound}
	svc := NewService(repo)

	err := svc.Delete("missing")
	if !errors.Is(err, core.ErrCompanyNotFound) {
		t.Fatalf("expected ErrCompanyNotFound, got %v", err)
	}
	if repo.deleted != "" {
		t.Fatalf("expected no delete call, got %s", repo.deleted)
	}
}
