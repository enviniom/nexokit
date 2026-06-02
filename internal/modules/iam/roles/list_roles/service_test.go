package list_roles

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type fakeRepo struct {
	items    []core.RoleResponse
	count    int64
	listErr  error
	countErr error
}

func (f fakeRepo) List(_ tenant.TenantContext, _ query.ListParams) ([]core.RoleResponse, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.items, nil
}

func (f fakeRepo) Count(_ tenant.TenantContext) (int64, error) {
	if f.countErr != nil {
		return 0, f.countErr
	}
	return f.count, nil
}

func TestListRoleService(t *testing.T) {
	tests := []struct {
		name      string
		repo      fakeRepo
		wantTotal int64
		wantLen   int
		wantErr   error
	}{
		{name: "success", repo: fakeRepo{items: []core.RoleResponse{{PublicID: "role-1"}}, count: 1}, wantTotal: 1, wantLen: 1},
		{name: "list error", repo: fakeRepo{listErr: errors.New("db list error")}, wantErr: errors.New("db list error")},
		{name: "count error", repo: fakeRepo{items: []core.RoleResponse{{PublicID: "role-1"}}, countErr: errors.New("db count error")}, wantErr: errors.New("db count error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			items, total, err := svc.List(tenant.NewRoot(), query.ListParams{})

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Fatalf("expected %d items, got %d", tt.wantLen, len(items))
			}
			if total != tt.wantTotal {
				t.Fatalf("expected total %d, got %d", tt.wantTotal, total)
			}
		})
	}
}
