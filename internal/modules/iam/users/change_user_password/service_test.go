package change_user_password

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
)

type fakeRepo struct {
	user        *core.IAMUser
	userErr     error
	rootRole    *core.IAMRole
	roleErr     error
	updateErr   error
	updatedID   uint
	updatedHash string
}

func (f *fakeRepo) GetByPublicID(tenant.TenantContext, string) (*core.IAMUser, error) {
	return f.user, f.userErr
}

func (f *fakeRepo) GetRoleBySlug(string) (*core.IAMRole, error) {
	return f.rootRole, f.roleErr
}

func (f *fakeRepo) UpdatePassword(userID uint, hash string) error {
	f.updatedID = userID
	f.updatedHash = hash
	return f.updateErr
}

type fakeHasher struct {
	hash      string
	hashErr   error
	verifyErr error
}

func (f fakeHasher) HashPassword(string) (string, error) {
	return f.hash, f.hashErr
}

func (f fakeHasher) VerifyPassword(_, _ string) error {
	return f.verifyErr
}

func rootRole() *core.IAMRole {
	return &core.IAMRole{BaseModel: shared.BaseModel{ID: 1, PublicID: "role-root"}, Slug: "root"}
}

func regularUser() *core.IAMUser {
	return &core.IAMUser{
		BaseModel:    shared.BaseModel{ID: 10, PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "old-hash",
		RoleID:       2,
		IsActive:     true,
	}
}

func rootUser() *core.IAMUser {
	return &core.IAMUser{
		BaseModel:    shared.BaseModel{ID: 20, PublicID: "user-root"},
		Name:         "Root",
		Email:        "root@example.com",
		PasswordHash: "root-hash",
		RoleID:       1,
		IsActive:     true,
	}
}

func TestChangeService(t *testing.T) {
	tests := []struct {
		name          string
		repo          *fakeRepo
		hasher        fakeHasher
		publicID      string
		actorPublicID string
		req           core.ChangePasswordRequest
		wantErr       error
		assertUpdated func(t *testing.T, repo *fakeRepo)
	}{
		{
			name: "success regular user",
			repo: &fakeRepo{
				user:     regularUser(),
				rootRole: rootRole(),
			},
			hasher:        fakeHasher{hash: "new-hash"},
			publicID:      "user-1",
			actorPublicID: "actor-1",
			req:           core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
			assertUpdated: func(t *testing.T, repo *fakeRepo) {
				if repo.updatedID != 10 {
					t.Fatalf("expected updated ID 10, got %d", repo.updatedID)
				}
				if repo.updatedHash != "new-hash" {
					t.Fatalf("expected updated hash new-hash, got %s", repo.updatedHash)
				}
			},
		},
		{
			name: "success root user changing own password",
			repo: &fakeRepo{
				user:     rootUser(),
				rootRole: rootRole(),
			},
			hasher:        fakeHasher{hash: "new-root-hash"},
			publicID:      "user-root",
			actorPublicID: "user-root",
			req:           core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
		},
		{
			name: "user not found",
			repo: &fakeRepo{
				userErr: core.ErrNotFound,
			},
			hasher:   fakeHasher{hash: "new-hash"},
			publicID: "missing",
			req:      core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
			wantErr:  core.ErrNotFound,
		},
		{
			name: "role lookup fails",
			repo: &fakeRepo{
				user:    regularUser(),
				roleErr: errors.New("db down"),
			},
			hasher:   fakeHasher{hash: "new-hash"},
			publicID: "user-1",
			req:      core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
			wantErr:  errors.New("db down"),
		},
		{
			name: "root user changed by different actor forbidden",
			repo: &fakeRepo{
				user:     rootUser(),
				rootRole: rootRole(),
			},
			hasher:        fakeHasher{hash: "new-hash"},
			publicID:      "user-root",
			actorPublicID: "other-actor",
			req:           core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
			wantErr:       core.ErrForbidden,
		},
		{
			name: "root user changed by anonymous actor forbidden",
			repo: &fakeRepo{
				user:     rootUser(),
				rootRole: rootRole(),
			},
			hasher:        fakeHasher{hash: "new-hash"},
			publicID:      "user-root",
			actorPublicID: "",
			req:           core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
			wantErr:       core.ErrForbidden,
		},
		{
			name: "wrong current password unauthorized",
			repo: &fakeRepo{
				user:     regularUser(),
				rootRole: rootRole(),
			},
			hasher:   fakeHasher{verifyErr: errors.New("wrong password")},
			publicID: "user-1",
			req:      core.ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "newpass123"},
			wantErr:  core.ErrUnauthorized,
		},
		{
			name: "hasher error",
			repo: &fakeRepo{
				user:     regularUser(),
				rootRole: rootRole(),
			},
			hasher:   fakeHasher{hashErr: errors.New("hash failed")},
			publicID: "user-1",
			req:      core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
			wantErr:  errors.New("hash failed"),
		},
		{
			name: "update error",
			repo: &fakeRepo{
				user:      regularUser(),
				rootRole:  rootRole(),
				updateErr: errors.New("update failed"),
			},
			hasher:   fakeHasher{hash: "new-hash"},
			publicID: "user-1",
			req:      core.ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"},
			wantErr:  errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, tt.hasher)
			err := svc.Change(tenant.NewRoot(), tt.publicID, tt.actorPublicID, tt.req)

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
			if tt.assertUpdated != nil {
				tt.assertUpdated(t, tt.repo)
			}
		})
	}
}
