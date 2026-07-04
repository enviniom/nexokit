package assign_role_to_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

// --- fake repository ---

type fakeRepo struct {
	user           *core.IAMUser
	rootRole       *core.IAMRole
	assignableRole *core.IAMRole
	assignResp     *core.UserResponse
	getUserErr     error
	roleSlugErr    error
	assignableErr  error
	assignErr      error
}

func (f fakeRepo) GetUserByPublicID(tenant.TenantContext, string) (*core.IAMUser, error) {
	return f.user, f.getUserErr
}

func (f fakeRepo) GetRoleBySlug(string) (*core.IAMRole, error) {
	return f.rootRole, f.roleSlugErr
}

func (f fakeRepo) GetAssignableRoleByPublicID(tenant.TenantContext, string) (*core.IAMRole, error) {
	return f.assignableRole, f.assignableErr
}

func (f fakeRepo) AssignRole(tenant.TenantContext, *core.IAMUser, uint) (*core.UserResponse, error) {
	return f.assignResp, f.assignErr
}

// --- helpers ---

func uintPtr(v uint) *uint { return &v }

func rootRole() *core.IAMRole {
	return &core.IAMRole{BaseModel: shared.BaseModel{ID: 99, PublicID: "role-root"}, Name: "Root", Slug: core.RootRoleSlug}
}

func regularUser() *core.IAMUser {
	return &core.IAMUser{
		BaseModel:    shared.BaseModel{ID: 10, PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       5,
		CompanyID:    uintPtr(1),
		IsActive:     true,
	}
}

func rootUser() *core.IAMUser {
	return &core.IAMUser{
		BaseModel:    shared.BaseModel{ID: 20, PublicID: "user-root"},
		Name:         "RootUser",
		Email:        "root@example.com",
		PasswordHash: "hash",
		RoleID:       99,
		IsActive:     true,
	}
}

func assignableRole() *core.IAMRole {
	return &core.IAMRole{BaseModel: shared.BaseModel{ID: 20, PublicID: "role-admin"}, Name: "Admin", Slug: "admin", CompanyID: uintPtr(1)}
}

func successResponse() *core.UserResponse {
	return &core.UserResponse{PublicID: "user-1", Name: "Alice", RoleID: 20, RoleName: "Admin"}
}

// --- table-driven tests ---

func TestChangeRole(t *testing.T) {
	tests := []struct {
		name       string
		repo       fakeRepo
		targetID   string
		actorID    string
		req        core.ChangeUserRoleRequest
		wantErr    error
		assertResp func(t *testing.T, resp *core.UserResponse)
	}{
		{
			name: "success",
			repo: fakeRepo{
				user:           regularUser(),
				rootRole:       rootRole(),
				assignableRole: assignableRole(),
				assignResp:     successResponse(),
			},
			targetID: "user-1",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			assertResp: func(t *testing.T, resp *core.UserResponse) {
				t.Helper()
				if resp.PublicID != "user-1" {
					t.Fatalf("expected user-1, got %s", resp.PublicID)
				}
				if resp.RoleID != 20 {
					t.Fatalf("expected role_id 20, got %d", resp.RoleID)
				}
			},
		},
		{
			name:     "empty actor forbidden",
			repo:     fakeRepo{},
			targetID: "user-1",
			actorID:  "",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			wantErr:  core.ErrForbidden,
		},
		{
			name:     "self assignment forbidden",
			repo:     fakeRepo{},
			targetID: "user-1",
			actorID:  "user-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			wantErr:  core.ErrForbidden,
		},
		{
			name: "user not found",
			repo: fakeRepo{
				getUserErr: core.ErrNotFound,
			},
			targetID: "missing",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			wantErr:  core.ErrNotFound,
		},
		{
			name: "target user is root forbidden",
			repo: fakeRepo{
				user:     rootUser(),
				rootRole: rootRole(),
			},
			targetID: "user-root",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			wantErr:  core.ErrForbidden,
		},
		{
			name: "role lookup error",
			repo: fakeRepo{
				user:        regularUser(),
				rootRole:    rootRole(),
				roleSlugErr: errors.New("db down"),
			},
			targetID: "user-1",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			wantErr:  errors.New("db down"),
		},
		{
			name: "assignable role not found",
			repo: fakeRepo{
				user:          regularUser(),
				rootRole:      rootRole(),
				assignableErr: core.ErrNotFound,
			},
			targetID: "user-1",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "missing-role"},
			wantErr:  core.ErrNotFound,
		},
		{
			name: "assign to root role forbidden",
			repo: fakeRepo{
				user:     regularUser(),
				rootRole: rootRole(),
				assignableRole: &core.IAMRole{
					BaseModel: shared.BaseModel{ID: 99, PublicID: "role-root"},
					Slug:      core.RootRoleSlug,
				},
			},
			targetID: "user-1",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-root"},
			wantErr:  core.ErrForbiddenRoleAssignment,
		},
		{
			name: "forbidden company scope mismatch",
			repo: fakeRepo{
				user:     regularUser(),
				rootRole: rootRole(),
				assignableRole: &core.IAMRole{
					BaseModel: shared.BaseModel{ID: 30, PublicID: "role-other"},
					Slug:      "other",
					CompanyID: uintPtr(999),
				},
			},
			targetID: "user-1",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-other"},
			wantErr:  core.ErrForbiddenCompanyScope,
		},
		{
			name: "assign persistence error",
			repo: fakeRepo{
				user:           regularUser(),
				rootRole:       rootRole(),
				assignableRole: assignableRole(),
				assignErr:      errors.New("db write failure"),
			},
			targetID: "user-1",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			wantErr:  errors.New("db write failure"),
		},
		{
			name: "user without company can be assigned to company role",
			repo: fakeRepo{
				user: &core.IAMUser{
					BaseModel:    shared.BaseModel{ID: 11, PublicID: "user-nocompany"},
					Name:         "Bob",
					Email:        "bob@example.com",
					PasswordHash: "hash",
					RoleID:       5,
					CompanyID:    nil,
					IsActive:     true,
				},
				rootRole:       rootRole(),
				assignableRole: assignableRole(),
				assignResp:     &core.UserResponse{PublicID: "user-nocompany", Name: "Bob", RoleID: 20, RoleName: "Admin"},
			},
			targetID: "user-nocompany",
			actorID:  "actor-1",
			req:      core.ChangeUserRoleRequest{RoleID: "role-admin"},
			assertResp: func(t *testing.T, resp *core.UserResponse) {
				t.Helper()
				if resp.PublicID != "user-nocompany" {
					t.Fatalf("expected user-nocompany, got %s", resp.PublicID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			resp, err := svc.ChangeRole(tenant.NewRoot(), tt.targetID, tt.actorID, tt.req)

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
			if tt.assertResp != nil {
				tt.assertResp(t, resp)
			}
		})
	}
}
