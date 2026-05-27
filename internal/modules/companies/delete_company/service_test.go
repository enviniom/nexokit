package delete_company

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/companies/core"
)

type fakeDeleteRepo struct{ deleted string }

func (f *fakeDeleteRepo) GetByPublicID(string) (*core.Company, error) {
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
