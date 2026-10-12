package list_selectable_roles

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
)

type fakeRepo struct {
	items []response.SelectResponse
	err   error
}

func (f fakeRepo) List(_ tenant.TenantContext) ([]response.SelectResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func TestServiceList(t *testing.T) {
	tests := []struct {
		name    string
		repo    fakeRepo
		wantLen int
		wantErr string
	}{
		{
			name:    "success returns selectable roles",
			repo:    fakeRepo{items: []response.SelectResponse{{ID: "role-1", Name: "Manager", Meta: map[string]any{"slug": "manager"}}}},
			wantLen: 1,
		},
		{
			name:    "repository error bubbles up",
			repo:    fakeRepo{err: errors.New("db down")},
			wantErr: "db down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			items, err := svc.List(tenant.NewRoot())

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Fatalf("expected %d items, got %d", tt.wantLen, len(items))
			}
		})
	}
}
