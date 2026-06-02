package update_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct {
	user       *core.IAMUser
	getByIDErr error

	rootRole    *core.IAMRole
	roleSlugErr error

	updateErr error

	reloaded  *core.UserResponse
	reloadErr error
}

func (f fakeRepo) GetByPublicID(tenant.TenantContext, string) (*core.IAMUser, error) {
	return f.user, f.getByIDErr
}

func (f fakeRepo) GetRoleBySlug(string) (*core.IAMRole, error) {
	return f.rootRole, f.roleSlugErr
}

func (f fakeRepo) Update(*core.IAMUser) error {
	return f.updateErr
}

func (f fakeRepo) Reload(tenant.TenantContext, string) (*core.UserResponse, error) {
	return f.reloaded, f.reloadErr
}

func rootRole() *core.IAMRole {
	return &core.IAMRole{BaseModel: shared.BaseModel{ID: 1, PublicID: "role-root"}, Slug: "root"}
}

func userRole() *core.IAMRole {
	return &core.IAMRole{BaseModel: shared.BaseModel{ID: 2, PublicID: "role-user"}, Slug: "user"}
}

func rootUser(roleID uint) *core.IAMUser {
	return &core.IAMUser{
		BaseModel:    shared.BaseModel{ID: 10, PublicID: "user-root"},
		Name:         "Root",
		Email:        "root@example.com",
		PasswordHash: "hash",
		RoleID:       roleID,
		IsActive:     true,
	}
}

func regularUser(roleID uint, companyID *uint) *core.IAMUser {
	return &core.IAMUser{
		BaseModel:    shared.BaseModel{ID: 20, PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       roleID,
		CompanyID:    companyID,
		IsActive:     true,
	}
}

func TestUpdateService(t *testing.T) {
	companyID := uint(10)

	tests := []struct {
		name          string
		repo          fakeRepo
		tc            tenant.TenantContext
		publicID      string
		actorPublicID string
		req           core.UpdateUserRequest
		wantErr       error
	}{
		{
			name: "success root user self-update",
			repo: fakeRepo{
				user:     rootUser(1),
				rootRole: rootRole(),
				reloaded: &core.UserResponse{PublicID: "user-root", Name: "Updated Root"},
			},
			tc:            tenant.NewRoot(),
			publicID:      "user-root",
			actorPublicID: "user-root",
			req:           core.UpdateUserRequest{Name: "Updated Root", Email: "root@example.com"},
		},
		{
			name: "success scoped user update",
			repo: fakeRepo{
				user:     regularUser(2, &companyID),
				rootRole: rootRole(),
				reloaded: &core.UserResponse{PublicID: "user-1", Name: "Updated Alice", CompanyID: &companyID},
			},
			tc:            tenant.NewScoped(companyID, "acme"),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "Updated Alice", Email: "alice@example.com"},
		},
		{
			name: "success root scope user update with company_id",
			repo: fakeRepo{
				user:     regularUser(2, &companyID),
				rootRole: rootRole(),
				reloaded: &core.UserResponse{PublicID: "user-1", Name: "Updated", CompanyID: &companyID},
			},
			tc:            tenant.NewRoot(),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "Updated", Email: "alice@example.com", CompanyID: &companyID},
		},
		{
			name: "not found",
			repo: fakeRepo{
				getByIDErr: core.ErrNotFound,
			},
			tc:            tenant.NewRoot(),
			publicID:      "missing",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "X", Email: "x@example.com"},
			wantErr:       core.ErrNotFound,
		},
		{
			name: "role slug lookup fails",
			repo: fakeRepo{
				user:        regularUser(2, &companyID),
				roleSlugErr: errors.New("db down"),
			},
			tc:            tenant.NewRoot(),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "X", Email: "x@example.com"},
			wantErr:       errors.New("db down"),
		},
		{
			name: "root user non-self actor forbidden",
			repo: fakeRepo{
				user:     rootUser(1),
				rootRole: rootRole(),
			},
			tc:            tenant.NewRoot(),
			publicID:      "user-root",
			actorPublicID: "other-user",
			req:           core.UpdateUserRequest{Name: "X", Email: "x@example.com"},
			wantErr:       core.ErrForbiddenRoleAssignment,
		},
		{
			name: "root user empty actor forbidden",
			repo: fakeRepo{
				user:     rootUser(1),
				rootRole: rootRole(),
			},
			tc:            tenant.NewRoot(),
			publicID:      "user-root",
			actorPublicID: "",
			req:           core.UpdateUserRequest{Name: "X", Email: "x@example.com"},
			wantErr:       core.ErrForbiddenRoleAssignment,
		},
		{
			name: "cross-company assignment forbidden",
			repo: fakeRepo{
				user:     regularUser(2, &companyID),
				rootRole: rootRole(),
			},
			tc:            tenant.NewScoped(companyID, "acme"),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "X", Email: "x@example.com", CompanyID: ptrUint(99)},
			wantErr:       core.ErrForbiddenRoleAssignment,
		},
		{
			name: "root scope without company_id invalid",
			repo: fakeRepo{
				user:     regularUser(2, &companyID),
				rootRole: rootRole(),
			},
			tc:            tenant.NewRoot(),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "X", Email: "x@example.com"},
			wantErr:       core.ErrInvalidCompanyScope,
		},
		{
			name: "duplicate email",
			repo: fakeRepo{
				user:      regularUser(2, &companyID),
				rootRole:  rootRole(),
				updateErr: core.ErrUserEmailAlreadyExists,
			},
			tc:            tenant.NewScoped(companyID, "acme"),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "Alice", Email: "taken@example.com"},
			wantErr:       core.ErrUserEmailAlreadyExists,
		},
		{
			name: "update error",
			repo: fakeRepo{
				user:      regularUser(2, &companyID),
				rootRole:  rootRole(),
				updateErr: errors.New("save failed"),
			},
			tc:            tenant.NewScoped(companyID, "acme"),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "Alice", Email: "alice@example.com"},
			wantErr:       errors.New("save failed"),
		},
		{
			name: "reload error",
			repo: fakeRepo{
				user:      regularUser(2, &companyID),
				rootRole:  rootRole(),
				reloadErr: errors.New("reload failed"),
			},
			tc:            tenant.NewScoped(companyID, "acme"),
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.UpdateUserRequest{Name: "Alice", Email: "alice@example.com"},
			wantErr:       errors.New("reload failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			resp, err := svc.Update(tt.tc, tt.publicID, tt.actorPublicID, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatalf("expected response, got nil")
			}
		})
	}
}

func ptrUint(v uint) *uint { return &v }
