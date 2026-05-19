package users

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// fakeRepository is a test double for the repository.
type fakeRepository struct {
	users          []User
	userByPublicID map[string]*User
	userByEmail    map[string]*User
	total          int64
	err            error
	getByEmailErr  error
	createErr      error
	updateErr      error
}

// fakeRoleResolver is a test double for the role resolver.
type fakeRoleResolver struct {
	role *roles.Role
	err  error
}

func (f *fakeRoleResolver) GetBySlug(slug string) (*roles.Role, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.role, nil
}

func (f *fakeRepository) List(page, perPage int) ([]User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.users, nil
}

func (f *fakeRepository) Count() (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.total, nil
}

func (f *fakeRepository) GetByPublicID(publicID string) (*User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.userByPublicID[publicID]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) GetByEmail(email string) (*User, error) {
	if f.getByEmailErr != nil {
		return nil, f.getByEmailErr
	}
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.userByEmail[email]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) Create(user *User) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.err != nil {
		return f.err
	}
	user.Role = roles.Role{Name: "admin"}
	f.users = append(f.users, *user)
	if f.userByPublicID == nil {
		f.userByPublicID = map[string]*User{}
	}
	copy := *user
	f.userByPublicID[user.PublicID] = &copy
	return nil
}

func (f *fakeRepository) Update(user *User) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if f.err != nil {
		return f.err
	}
	user.Role = roles.Role{Name: "admin"}
	for i := range f.users {
		if f.users[i].PublicID == user.PublicID {
			f.users[i] = *user
			if f.userByPublicID == nil {
				f.userByPublicID = map[string]*User{}
			}
			copy := *user
			f.userByPublicID[user.PublicID] = &copy
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeRepository) Delete(publicID string) error {
	if f.err != nil {
		return f.err
	}
	for i := range f.users {
		if f.users[i].PublicID == publicID {
			f.users = append(f.users[:i], f.users[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeRepository) ListPublicIDsByRoleID(roleID uint) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	ids := make([]string, 0)
	for _, user := range f.users {
		if user.RoleID == roleID {
			ids = append(ids, user.PublicID)
		}
	}
	return ids, nil
}

// fakeHasher is a test double for the password hasher.
type fakeHasher struct {
	hash string
	err  error
}

func (f *fakeHasher) HashPassword(password string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.hash, nil
}

func (f *fakeHasher) VerifyPassword(password, hash string) error {
	if f.err != nil {
		return f.err
	}
	if password == "wrong" {
		return errors.New("password does not match")
	}
	return nil
}

func TestService_List(t *testing.T) {
	t.Run("returns paginated users", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1, Role: roles.Role{Name: "admin"}},
				{BaseModel: shared.BaseModel{PublicID: "user2"}, Name: "Bob", Email: "bob@example.com", IsActive: true, RoleID: 2, Role: roles.Role{Name: "user"}},
			},
			total: 2,
		}
		svc := NewService(repo, &fakeHasher{hash: "fakehash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		result, total, err := svc.List(1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 users, got %d", len(result))
		}
		if result[0].Name != "Alice" {
			t.Errorf("expected first user name 'Alice', got %s", result[0].Name)
		}
		if result[0].RoleName != "admin" {
			t.Errorf("expected first user role_name 'admin', got %s", result[0].RoleName)
		}
	})

	t.Run("returns empty list when no users", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, total: 0}
		svc := NewService(repo, &fakeHasher{hash: "fakehash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		result, total, err := svc.List(1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 users, got %d", len(result))
		}
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		repo := &fakeRepository{err: apperror.ErrInternal}
		svc := NewService(repo, &fakeHasher{hash: "fakehash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		_, _, err := svc.List(1, 10)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestService_GetByPublicID(t *testing.T) {
	t.Run("returns user when found", func(t *testing.T) {
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1, Role: roles.Role{Name: "admin"}},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "fakehash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		result, err := svc.GetByPublicID("user1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.PublicID != "user1" {
			t.Errorf("expected public_id 'user1', got %s", result.PublicID)
		}
		if result.RoleName != "admin" {
			t.Errorf("expected role_name 'admin', got %s", result.RoleName)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &fakeRepository{userByPublicID: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "fakehash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		_, err := svc.GetByPublicID("missing")
		if err == nil {
			t.Error("expected error for missing user")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestService_Create(t *testing.T) {
	t.Run("creates a new user successfully", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 1}
		result, err := svc.Create(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "Alice" {
			t.Errorf("expected name 'Alice', got %s", result.Name)
		}
		if result.Email != "alice@example.com" {
			t.Errorf("expected email 'alice@example.com', got %s", result.Email)
		}
		if result.RoleName != "admin" {
			t.Errorf("expected role_name 'admin', got %s", result.RoleName)
		}
	})

	t.Run("returns conflict when email already exists", func(t *testing.T) {
		repo := &fakeRepository{
			userByEmail: map[string]*User{
				"alice@example.com": {BaseModel: shared.BaseModel{PublicID: "u1"}, Email: "alice@example.com"},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 1}
		_, err := svc.Create(req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns repository error when email uniqueness check fails", func(t *testing.T) {
		repo := &fakeRepository{getByEmailErr: apperror.ErrInternal}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 1}
		_, err := svc.Create(req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrInternal) {
			t.Errorf("expected ErrInternal, got %v", err)
		}
	})

	t.Run("returns conflict when repository create hits unique constraint", func(t *testing.T) {
		repo := &fakeRepository{userByEmail: map[string]*User{}, createErr: gorm.ErrDuplicatedKey}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 1}
		_, err := svc.Create(req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestService_Create_RejectsRootRole(t *testing.T) {
	t.Run("returns forbidden when creating user with root role", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 1}, Name: "root"}})

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 1}
		_, err := svc.Create(req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestService_Update(t *testing.T) {
	t.Run("updates a user successfully", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			userByEmail: map[string]*User{},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := UpdateUserRequest{Name: "Alice Updated", Email: "alice-new@example.com", RoleID: 2}
		result, err := svc.Update("user1", "", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "Alice Updated" {
			t.Errorf("expected name 'Alice Updated', got %s", result.Name)
		}
		if result.Email != "alice-new@example.com" {
			t.Errorf("expected email 'alice-new@example.com', got %s", result.Email)
		}
		if result.RoleName != "admin" {
			t.Errorf("expected role_name 'admin', got %s", result.RoleName)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &fakeRepository{userByPublicID: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		_, err := svc.Update("missing", "", UpdateUserRequest{Name: "Alice", Email: "alice@example.com", RoleID: 1})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("returns conflict when email already exists", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
				{BaseModel: shared.BaseModel{PublicID: "user2"}, Name: "Bob", Email: "bob@example.com", RoleID: 2},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			userByEmail: map[string]*User{
				"bob@example.com": {BaseModel: shared.BaseModel{PublicID: "user2"}, Email: "bob@example.com"},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := UpdateUserRequest{Name: "Alice", Email: "bob@example.com", RoleID: 1}
		_, err := svc.Update("user1", "", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns repository error when email uniqueness check fails", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			getByEmailErr: apperror.ErrInternal,
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := UpdateUserRequest{Name: "Alice", Email: "bob@example.com", RoleID: 1}
		_, err := svc.Update("user1", "", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrInternal) {
			t.Errorf("expected ErrInternal, got %v", err)
		}
	})

	t.Run("returns conflict when repository update hits unique constraint", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			userByEmail: map[string]*User{},
			updateErr:   gorm.ErrDuplicatedKey,
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := UpdateUserRequest{Name: "Alice", Email: "bob@example.com", RoleID: 1}
		_, err := svc.Update("user1", "", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns forbidden when promoting non-root to root", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 2},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 2},
			},
			userByEmail: map[string]*User{},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 1}, Name: "root"}})

		req := UpdateUserRequest{Name: "Alice", Email: "alice@example.com", RoleID: 1}
		_, err := svc.Update("user1", "", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns forbidden when editing root without actor", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "root1"}, Name: "Root", Email: "root@example.com", RoleID: 1},
			},
			userByPublicID: map[string]*User{
				"root1": {BaseModel: shared.BaseModel{PublicID: "root1"}, Name: "Root", Email: "root@example.com", RoleID: 1},
			},
			userByEmail: map[string]*User{},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 1}, Name: "root"}})

		req := UpdateUserRequest{Name: "Root Updated", Email: "root-new@example.com", RoleID: 1}
		_, err := svc.Update("root1", "", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns forbidden when editing root by different actor", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "root1"}, Name: "Root", Email: "root@example.com", RoleID: 1},
			},
			userByPublicID: map[string]*User{
				"root1": {BaseModel: shared.BaseModel{PublicID: "root1"}, Name: "Root", Email: "root@example.com", RoleID: 1},
			},
			userByEmail: map[string]*User{},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 1}, Name: "root"}})

		req := UpdateUserRequest{Name: "Root Updated", Email: "root-new@example.com", RoleID: 1}
		_, err := svc.Update("root1", "other", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("allows root self-edit and only updates name and email", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "root1"}, Name: "Root", Email: "root@example.com", RoleID: 1, CompanyID: nil},
			},
			userByPublicID: map[string]*User{
				"root1": {BaseModel: shared.BaseModel{PublicID: "root1"}, Name: "Root", Email: "root@example.com", RoleID: 1, CompanyID: nil},
			},
			userByEmail: map[string]*User{},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 1}, Name: "root"}})

		companyID := uint(42)
		req := UpdateUserRequest{Name: "Root Updated", Email: "root-new@example.com", RoleID: 2, CompanyID: &companyID}
		result, err := svc.Update("root1", "root1", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "Root Updated" {
			t.Errorf("expected name 'Root Updated', got %s", result.Name)
		}
		if result.Email != "root-new@example.com" {
			t.Errorf("expected email 'root-new@example.com', got %s", result.Email)
		}
		// RoleID and CompanyID should be ignored for root self-edit
		if result.RoleID != 1 {
			t.Errorf("expected role_id to remain 1, got %d", result.RoleID)
		}
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("deletes a user successfully", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 1},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		err := svc.Delete("user1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &fakeRepository{userByPublicID: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		err := svc.Delete("missing")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestService_ChangePassword(t *testing.T) {
	t.Run("changes password with correct current password", func(t *testing.T) {
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Email: "alice@example.com", PasswordHash: "oldhash"},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Email: "alice@example.com", PasswordHash: "oldhash"},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "newhash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		err := svc.ChangePassword("user1", "", ChangePasswordRequest{CurrentPassword: "oldpassword", NewPassword: "NewPassword1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.users[0].PasswordHash != "newhash" {
			t.Errorf("expected password hash to be updated to 'newhash', got %s", repo.users[0].PasswordHash)
		}
	})

	t.Run("returns unauthorized with wrong current password", func(t *testing.T) {
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Email: "alice@example.com", PasswordHash: "oldhash"},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "newhash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		err := svc.ChangePassword("user1", "", ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "NewPassword1"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("returns not found when user missing", func(t *testing.T) {
		repo := &fakeRepository{userByPublicID: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "newhash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		err := svc.ChangePassword("missing", "", ChangePasswordRequest{CurrentPassword: "old", NewPassword: "NewPassword1"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestService_ToggleStatus(t *testing.T) {
	t.Run("toggles user status to inactive", func(t *testing.T) {
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1},
			},
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		result, err := svc.ToggleStatus("user1", UpdateStatusRequest{IsActive: false})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsActive {
			t.Errorf("expected user to be inactive")
		}
	})

	t.Run("toggles user status to active", func(t *testing.T) {
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", IsActive: false, RoleID: 1},
			},
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", IsActive: false, RoleID: 1},
			},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		result, err := svc.ToggleStatus("user1", UpdateStatusRequest{IsActive: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsActive {
			t.Errorf("expected user to be active")
		}
	})

	t.Run("returns not found when user missing", func(t *testing.T) {
		repo := &fakeRepository{userByPublicID: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		_, err := svc.ToggleStatus("missing", UpdateStatusRequest{IsActive: false})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
