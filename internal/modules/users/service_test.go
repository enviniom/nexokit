package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

const RootRoleID uint = 1

func uintPtr(v uint) *uint { return &v }

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
	lastTenant     tenant.TenantContext
	listParams     query.ListParams
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

func (f *fakeRepository) List(tc tenant.TenantContext, params query.ListParams) ([]User, error) {
	f.lastTenant = tc
	f.listParams = params
	if f.err != nil {
		return nil, f.err
	}
	return filterUsersByTenant(f.users, tc), nil
}

func (f *fakeRepository) Count(tc tenant.TenantContext, params query.ListParams) (int64, error) {
	f.lastTenant = tc
	f.listParams = params
	if f.err != nil {
		return 0, f.err
	}
	if f.total != 0 {
		return f.total, nil
	}
	return int64(len(filterUsersByTenant(f.users, tc))), nil
}

func (f *fakeRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*User, error) {
	f.lastTenant = tc
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.userByPublicID[publicID]; ok {
		if !tenantAllowsUser(tc, u) {
			return nil, gorm.ErrRecordNotFound
		}
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepository) GetAuthUser(publicID string) (*User, error) {
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

func (f *fakeRepository) Delete(tc tenant.TenantContext, publicID string) error {
	f.lastTenant = tc
	if f.err != nil {
		return f.err
	}
	for i := range f.users {
		if f.users[i].PublicID == publicID && tenantAllowsUser(tc, &f.users[i]) {
			f.users = append(f.users[:i], f.users[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func filterUsersByTenant(users []User, tc tenant.TenantContext) []User {
	if tc.IsRootScope {
		return users
	}
	filtered := make([]User, 0, len(users))
	for _, user := range users {
		if tenantAllowsUser(tc, &user) {
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func tenantAllowsUser(tc tenant.TenantContext, user *User) bool {
	return tc.IsRootScope || (user.CompanyID != nil && *user.CompanyID == tc.CompanyID)
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

func (f *fakeRepository) CountByRoleID(roleID uint) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	var count int64
	for _, user := range f.users {
		if user.RoleID == roleID {
			count++
		}
	}
	return count, nil
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

		params := query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}}
		result, total, err := svc.List(tenant.NewRoot(), params)
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

		result, total, err := svc.List(tenant.NewRoot(), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
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

		_, _, err := svc.List(tenant.NewRoot(), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("passes filters search sorting and dates to repository", func(t *testing.T) {
		repo := &fakeRepository{users: []User{{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", IsActive: true, Role: roles.Role{Name: "admin"}}}, total: 1}
		svc := NewService(repo, &fakeHasher{hash: "fakehash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})
		from := mustDate(t, "2025-01-01")
		to := mustDate(t, "2025-12-31")
		params := query.ListParams{
			Pagination: query.PaginationParams{Page: 2, PerPage: 5},
			Filters:    query.FilterParams{Status: "active", CreatedFrom: &from, CreatedTo: &to},
			Sort:       query.SortParams{Sort: "email", Order: "asc"},
			Search:     query.SearchParams{Query: "alice"},
		}

		result, total, err := svc.List(tenant.NewRoot(), params)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 || len(result) != 1 || result[0].Email != "alice@example.com" {
			t.Fatalf("expected Alice result, total=%d result=%#v", total, result)
		}
		if repo.listParams.Sort.Sort != "email" || repo.listParams.Search.Query != "alice" || repo.listParams.Filters.Status != "active" {
			t.Fatalf("expected repository to receive list params, got %#v", repo.listParams)
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

		result, err := svc.GetByPublicID(tenant.NewRoot(), "user1")
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

		_, err := svc.GetByPublicID(tenant.NewRoot(), "missing")
		if err == nil {
			t.Error("expected error for missing user")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestService_TenantScopedReads(t *testing.T) {
	companyOne := uint(1)
	companyTwo := uint(2)
	repo := &fakeRepository{
		users: []User{
			{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", CompanyID: &companyOne, Role: roles.Role{Name: "admin"}},
			{BaseModel: shared.BaseModel{PublicID: "user2"}, Name: "Bob", Email: "bob@example.com", CompanyID: &companyTwo, Role: roles.Role{Name: "admin"}},
			{BaseModel: shared.BaseModel{PublicID: "root1"}, Name: "Root", Email: "root@example.com", CompanyID: nil, Role: roles.Role{Name: "root"}},
		},
		userByPublicID: map[string]*User{
			"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", CompanyID: &companyOne, Role: roles.Role{Name: "admin"}},
			"user2": {BaseModel: shared.BaseModel{PublicID: "user2"}, Name: "Bob", Email: "bob@example.com", CompanyID: &companyTwo, Role: roles.Role{Name: "admin"}},
		},
	}
	svc := NewService(repo, &fakeHasher{hash: "fakehash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}})

	t.Run("admin sees only own company users", func(t *testing.T) {
		result, total, err := svc.List(tenant.NewScoped(companyOne, "acme"), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 || len(result) != 1 || result[0].PublicID != "user1" {
			t.Fatalf("expected only company one user, total=%d result=%#v", total, result)
		}
	})

	t.Run("root global sees all users", func(t *testing.T) {
		result, total, err := svc.List(tenant.NewRoot(), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 || len(result) != 3 {
			t.Fatalf("expected all users for root, total=%d len=%d", total, len(result))
		}
	})

	t.Run("root scoped sees one company", func(t *testing.T) {
		result, total, err := svc.List(tenant.NewScoped(companyTwo, "globex"), query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 || len(result) != 1 || result[0].PublicID != "user2" {
			t.Fatalf("expected only company two user, total=%d result=%#v", total, result)
		}
	})

	t.Run("cross tenant get returns not found", func(t *testing.T) {
		_, err := svc.GetByPublicID(tenant.NewScoped(companyOne, "acme"), "user2")
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for cross-tenant read, got %v", err)
		}
	})
}

func TestService_Create(t *testing.T) {
	t.Run("creates a new user successfully", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(1)}
		result, err := svc.Create(tenant.NewRoot(), req)
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

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(1)}
		_, err := svc.Create(tenant.NewRoot(), req)
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

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(1)}
		_, err := svc.Create(tenant.NewRoot(), req)
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

		req := CreateUserRequest{Name: "Alice", Email: "alice@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(1)}
		_, err := svc.Create(tenant.NewRoot(), req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestService_Create_AllowsRootWithoutCompany(t *testing.T) {
	t.Run("creates root with nullable company", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}})

		req := CreateUserRequest{Name: "Root", Email: "root@example.com", Password: "Password1", RoleID: RootRoleID}
		result, err := svc.Create(tenant.NewRoot(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.CompanyID != nil {
			t.Fatalf("expected root company_id to stay nil, got %v", *result.CompanyID)
		}
	})
}

func TestService_Create_TenantIsolation(t *testing.T) {
	rootRole := &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}

	t.Run("root global requires company for non-root user", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: rootRole})

		_, err := svc.Create(tenant.NewRoot(), CreateUserRequest{Name: "Admin", Email: "admin@example.com", Password: "Password1", RoleID: 2})

		if !errors.Is(err, apperror.ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest, got %v", err)
		}
	})

	t.Run("scoped tenant forces missing company to current tenant", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: rootRole})

		result, err := svc.Create(tenant.NewScoped(7, "acme"), CreateUserRequest{Name: "Admin", Email: "admin@example.com", Password: "Password1", RoleID: 2})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.CompanyID == nil || *result.CompanyID != 7 {
			t.Fatalf("expected company_id 7, got %v", result.CompanyID)
		}
	})

	t.Run("scoped tenant rejects another company", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: rootRole})

		_, err := svc.Create(tenant.NewScoped(7, "acme"), CreateUserRequest{Name: "Admin", Email: "admin@example.com", Password: "Password1", RoleID: 2, CompanyID: uintPtr(8)})

		if !errors.Is(err, apperror.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("only root global can create root without company", func(t *testing.T) {
		repo := &fakeRepository{users: []User{}, userByEmail: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: rootRole})

		_, err := svc.Create(tenant.NewScoped(7, "acme"), CreateUserRequest{Name: "Root", Email: "root@example.com", Password: "Password1", RoleID: RootRoleID})

		if !errors.Is(err, apperror.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
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

		req := UpdateUserRequest{Name: "Alice Updated", Email: "alice-new@example.com", CompanyID: uintPtr(1)}
		result, err := svc.Update(tenant.NewRoot(), "user1", "", req)
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

		_, err := svc.Update(tenant.NewRoot(), "missing", "", UpdateUserRequest{Name: "Alice", Email: "alice@example.com", CompanyID: uintPtr(1)})
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

		req := UpdateUserRequest{Name: "Alice", Email: "bob@example.com", CompanyID: uintPtr(1)}
		_, err := svc.Update(tenant.NewRoot(), "user1", "", req)
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

		req := UpdateUserRequest{Name: "Alice", Email: "bob@example.com", CompanyID: uintPtr(1)}
		_, err := svc.Update(tenant.NewRoot(), "user1", "", req)
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

		req := UpdateUserRequest{Name: "Alice", Email: "bob@example.com", CompanyID: uintPtr(1)}
		_, err := svc.Update(tenant.NewRoot(), "user1", "", req)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("asserts RoleID cannot be changed via general update", func(t *testing.T) {
		companyID := uint(1)
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 2, CompanyID: &companyID},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 2, CompanyID: &companyID},
			},
			userByEmail: map[string]*User{},
		}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 1}, Name: "root"}})

		req := UpdateUserRequest{Name: "Alice", Email: "alice@example.com", CompanyID: &companyID}
		result, err := svc.Update(tenant.NewRoot(), "user1", "", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RoleID != 2 {
			t.Errorf("expected RoleID to remain 2, got %d", result.RoleID)
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

		req := UpdateUserRequest{Name: "Root Updated", Email: "root-new@example.com"}
		_, err := svc.Update(tenant.NewRoot(), "root1", "", req)
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

		req := UpdateUserRequest{Name: "Root Updated", Email: "root-new@example.com"}
		_, err := svc.Update(tenant.NewRoot(), "root1", "other", req)
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
		req := UpdateUserRequest{Name: "Root Updated", Email: "root-new@example.com", CompanyID: &companyID}
		result, err := svc.Update(tenant.NewRoot(), "root1", "root1", req)
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

func TestService_TenantScopedWrites(t *testing.T) {
	companyOne := uint(1)
	companyTwo := uint(2)
	repo := &fakeRepository{
		users: []User{
			{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", PasswordHash: "oldhash", RoleID: 2, CompanyID: &companyOne, IsActive: true},
			{BaseModel: shared.BaseModel{PublicID: "user2"}, Name: "Bob", Email: "bob@example.com", PasswordHash: "oldhash", RoleID: 2, CompanyID: &companyTwo, IsActive: true},
		},
		userByPublicID: map[string]*User{
			"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", PasswordHash: "oldhash", RoleID: 2, CompanyID: &companyOne, IsActive: true},
			"user2": {BaseModel: shared.BaseModel{PublicID: "user2"}, Name: "Bob", Email: "bob@example.com", PasswordHash: "oldhash", RoleID: 2, CompanyID: &companyTwo, IsActive: true},
		},
		userByEmail: map[string]*User{},
	}
	svc := NewService(repo, &fakeHasher{hash: "newhash"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}})
	scope := tenant.NewScoped(companyOne, "acme")

	t.Run("updates user within tenant scope", func(t *testing.T) {
		result, err := svc.Update(scope, "user1", "", UpdateUserRequest{Name: "Alice Updated", Email: "alice-new@example.com", CompanyID: &companyOne})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "Alice Updated" || result.CompanyID == nil || *result.CompanyID != companyOne {
			t.Fatalf("expected updated company-one user, got %#v", result)
		}
	})

	t.Run("cross tenant update returns not found", func(t *testing.T) {
		_, err := svc.Update(scope, "user2", "", UpdateUserRequest{Name: "Bob Updated", Email: "bob-new@example.com", CompanyID: &companyTwo})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for cross-tenant update, got %v", err)
		}
	})

	t.Run("cross tenant delete returns not found", func(t *testing.T) {
		err := svc.Delete(scope, "user2")
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for cross-tenant delete, got %v", err)
		}
	})

	t.Run("cross tenant password change returns not found", func(t *testing.T) {
		err := svc.ChangePassword(scope, "user2", "", ChangePasswordRequest{CurrentPassword: "oldpassword", NewPassword: "NewPassword1"})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for cross-tenant password change, got %v", err)
		}
	})

	t.Run("cross tenant status toggle returns not found", func(t *testing.T) {
		_, err := svc.ToggleStatus(scope, "user2", UpdateStatusRequest{IsActive: false})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for cross-tenant status toggle, got %v", err)
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

		err := svc.Delete(tenant.NewRoot(), "user1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		repo := &fakeRepository{userByPublicID: map[string]*User{}}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: 999}, Name: "root"}})

		err := svc.Delete(tenant.NewRoot(), "missing")
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

		err := svc.ChangePassword(tenant.NewRoot(), "user1", "", ChangePasswordRequest{CurrentPassword: "oldpassword", NewPassword: "NewPassword1"})
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

		err := svc.ChangePassword(tenant.NewRoot(), "user1", "", ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "NewPassword1"})
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

		err := svc.ChangePassword(tenant.NewRoot(), "missing", "", ChangePasswordRequest{CurrentPassword: "old", NewPassword: "NewPassword1"})
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

		result, err := svc.ToggleStatus(tenant.NewRoot(), "user1", UpdateStatusRequest{IsActive: false})
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

		result, err := svc.ToggleStatus(tenant.NewRoot(), "user1", UpdateStatusRequest{IsActive: true})
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

		_, err := svc.ToggleStatus(tenant.NewRoot(), "missing", UpdateStatusRequest{IsActive: false})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

type fakeRoleReader struct {
	summary *roles.AssignmentRoleSummary
	err     error
}

func (f *fakeRoleReader) FindAssignableByPublicID(tc tenant.TenantContext, publicID string) (*roles.AssignmentRoleSummary, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.summary, nil
}

type fakeCache struct {
	deleted []string
}

func (f *fakeCache) Get(ctx context.Context, key string) ([]byte, error) { return nil, nil }
func (f *fakeCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}
func (f *fakeCache) Delete(ctx context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return nil
}
func (f *fakeCache) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (f *fakeCache) Close() error                                         { return nil }

func TestService_ChangeRole(t *testing.T) {
	t.Run("successfully assigns role and invalidates cache", func(t *testing.T) {
		companyOne := uint(1)
		repo := &fakeRepository{
			users: []User{
				{BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 2, CompanyID: &companyOne},
			},
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, Name: "Alice", Email: "alice@example.com", RoleID: 2, CompanyID: &companyOne},
			},
		}
		roleReader := &fakeRoleReader{
			summary: &roles.AssignmentRoleSummary{InternalID: 3, PublicID: "role3", Slug: "editor", CompanyID: &companyOne},
		}
		cacheMock := &fakeCache{}
		svc := NewService(repo, &fakeHasher{hash: "hashed"}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}}, WithRoleReader(roleReader), WithCache(cacheMock))

		req := ChangeUserRoleRequest{RoleID: "role3"}
		result, err := svc.ChangeRole(tenant.NewScoped(companyOne, "acme"), "user1", "admin1", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RoleID != 3 {
			t.Errorf("expected new role ID to be 3, got %d", result.RoleID)
		}
		if len(cacheMock.deleted) != 1 || cacheMock.deleted[0] != "rbac:permissions:user1" {
			t.Errorf("expected permission cache to be invalidated for target user, got %v", cacheMock.deleted)
		}
	})

	t.Run("rejects empty actor public ID", func(t *testing.T) {
		svc := NewService(&fakeRepository{}, &fakeHasher{}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}})
		_, err := svc.ChangeRole(tenant.NewRoot(), "user1", "", ChangeUserRoleRequest{RoleID: "role3"})
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden for empty actor, got %v", err)
		}
	})

	t.Run("rejects self role change", func(t *testing.T) {
		svc := NewService(&fakeRepository{}, &fakeHasher{}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}})
		_, err := svc.ChangeRole(tenant.NewRoot(), "user1", "user1", ChangeUserRoleRequest{RoleID: "role3"})
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden for self role change, got %v", err)
		}
	})

	t.Run("returns not found if target user missing or out of tenant scope", func(t *testing.T) {
		companyOne := uint(1)
		companyTwo := uint(2)
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, CompanyID: &companyTwo},
			},
		}
		svc := NewService(repo, &fakeHasher{}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}})
		_, err := svc.ChangeRole(tenant.NewScoped(companyOne, "acme"), "user1", "admin1", ChangeUserRoleRequest{RoleID: "role3"})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound for cross-tenant target user, got %v", err)
		}
	})

	t.Run("returns not found if role is missing or out of tenant scope", func(t *testing.T) {
		companyOne := uint(1)
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, CompanyID: &companyOne},
			},
		}
		roleReader := &fakeRoleReader{err: gorm.ErrRecordNotFound}
		svc := NewService(repo, &fakeHasher{}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}}, WithRoleReader(roleReader))
		_, err := svc.ChangeRole(tenant.NewScoped(companyOne, "acme"), "user1", "admin1", ChangeUserRoleRequest{RoleID: "role3"})
		if !errors.Is(err, apperror.ErrNotFound) {
			t.Errorf("expected ErrNotFound for missing role, got %v", err)
		}
	})

	t.Run("rejects root role assignment", func(t *testing.T) {
		companyOne := uint(1)
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, CompanyID: &companyOne},
			},
		}
		roleReader := &fakeRoleReader{
			summary: &roles.AssignmentRoleSummary{InternalID: 1, PublicID: "root-role-id", Slug: roles.RootRoleSlug},
		}
		svc := NewService(repo, &fakeHasher{}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}}, WithRoleReader(roleReader))
		_, err := svc.ChangeRole(tenant.NewRoot(), "user1", "admin1", ChangeUserRoleRequest{RoleID: "root-role-id"})
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden when assigning root role, got %v", err)
		}
	})

	t.Run("rejects role/user company mismatch", func(t *testing.T) {
		companyOne := uint(1)
		companyTwo := uint(2)
		repo := &fakeRepository{
			userByPublicID: map[string]*User{
				"user1": {BaseModel: shared.BaseModel{PublicID: "user1"}, CompanyID: &companyOne},
			},
		}
		roleReader := &fakeRoleReader{
			summary: &roles.AssignmentRoleSummary{InternalID: 3, PublicID: "role3", Slug: "editor", CompanyID: &companyTwo},
		}
		svc := NewService(repo, &fakeHasher{}, &fakeRoleResolver{role: &roles.Role{BaseModel: shared.BaseModel{ID: RootRoleID}, Name: "root"}}, WithRoleReader(roleReader))
		_, err := svc.ChangeRole(tenant.NewRoot(), "user1", "admin1", ChangeUserRoleRequest{RoleID: "role3"})
		if !errors.Is(err, apperror.ErrForbidden) {
			t.Errorf("expected ErrForbidden when target user company and role company mismatch, got %v", err)
		}
	})
}
