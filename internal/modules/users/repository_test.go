package users

import (
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/platform/query"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepository_ListAppliesFiltersSearchSortingAndPagination(t *testing.T) {
	db := newUserRepositoryTestDB(t)
	role := roles.Role{BaseModel: shared.BaseModel{PublicID: "role_admin"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	companyOne := uint(1)
	companyTwo := uint(2)
	createdFrom := mustDate(t, "2025-01-01")
	createdTo := mustDate(t, "2025-12-31")
	users := []User{
		{BaseModel: shared.BaseModel{PublicID: "user_b", CreatedAt: mustDate(t, "2025-03-02")}, Name: "Beta", Email: "beta@example.com", PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyOne, IsActive: true},
		{BaseModel: shared.BaseModel{PublicID: "user_a", CreatedAt: mustDate(t, "2025-03-01")}, Name: "Alpha", Email: "alpha@example.com", PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyOne, IsActive: true},
		{BaseModel: shared.BaseModel{PublicID: "user_inactive", CreatedAt: mustDate(t, "2025-03-03")}, Name: "Inactive", Email: "inactive@example.com", PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyOne, IsActive: false},
		{BaseModel: shared.BaseModel{PublicID: "user_other", CreatedAt: mustDate(t, "2025-03-04")}, Name: "Alpha Other", Email: "other@example.com", PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyTwo, IsActive: true},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	if err := db.Model(&User{}).Where("public_id = ?", "user_inactive").Update("is_active", false).Error; err != nil {
		t.Fatalf("mark inactive user: %v", err)
	}

	repo := NewRepository(db)
	params := query.ListParams{
		Pagination: query.PaginationParams{Page: 1, PerPage: 10},
		Filters:    query.FilterParams{Status: "active", CreatedFrom: &createdFrom, CreatedTo: &createdTo},
		Sort:       query.SortParams{Sort: "name", Order: "asc"},
		Search:     query.SearchParams{Query: "a"},
	}

	got, err := repo.List(tenant.NewScoped(companyOne, "acme"), params)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	total, err := repo.Count(tenant.NewScoped(companyOne, "acme"), params)
	if err != nil {
		t.Fatalf("count users: %v", err)
	}

	if total != 2 {
		t.Fatalf("expected filtered count 2, got %d", total)
	}
	if len(got) != 2 || got[0].PublicID != "user_a" || got[1].PublicID != "user_b" {
		t.Fatalf("expected active company users sorted by name, got %#v", got)
	}
}

func TestGormRepository_ListUsesExplicitSortAllowlist(t *testing.T) {
	db := newUserRepositoryTestDB(t)
	role := roles.Role{BaseModel: shared.BaseModel{PublicID: "role_admin"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	companyID := uint(1)
	users := []User{
		{BaseModel: shared.BaseModel{PublicID: "user_new", CreatedAt: mustDate(t, "2025-03-02")}, Name: "Zeta", Email: "zeta@example.com", PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyID, IsActive: true},
		{BaseModel: shared.BaseModel{PublicID: "user_old", CreatedAt: mustDate(t, "2025-03-01")}, Name: "Alpha", Email: "alpha@example.com", PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyID, IsActive: true},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	repo := NewRepository(db)
	params := query.ListParams{Pagination: query.PaginationParams{Page: 1, PerPage: 10}, Sort: query.SortParams{Sort: "password_hash", Order: "asc"}}

	got, err := repo.List(tenant.NewScoped(companyID, "acme"), params)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if len(got) != 2 || got[0].PublicID != "user_new" || got[1].PublicID != "user_old" {
		t.Fatalf("expected disallowed sort to fall back to created_at desc, got %#v", got)
	}
}

func TestGormRepository_DeleteSoftDeletesUsers(t *testing.T) {
	db := newUserRepositoryTestDB(t)
	role := roles.Role{BaseModel: shared.BaseModel{PublicID: "role_admin"}, Name: "Admin", Slug: "admin"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	companyID := uint(1)
	user := User{BaseModel: shared.BaseModel{PublicID: "user_delete"}, Name: "Delete Me", Email: "delete@example.com", PasswordHash: "hash", RoleID: role.ID, CompanyID: &companyID, IsActive: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewRepository(db)

	if err := repo.Delete(tenant.NewScoped(companyID, "acme"), "user_delete"); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	if _, err := repo.GetByPublicID(tenant.NewScoped(companyID, "acme"), "user_delete"); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected normal read to exclude soft-deleted user, got %v", err)
	}
	var deleted User
	if err := db.Unscoped().Where("public_id = ?", "user_delete").First(&deleted).Error; err != nil {
		t.Fatalf("expected soft-deleted row to remain: %v", err)
	}
	if !deleted.DeletedAt.Valid {
		t.Fatal("expected deleted_at to be set")
	}
}

func TestGormRepository_CountByRoleID(t *testing.T) {
	db := newUserRepositoryTestDB(t)
	roleOne := roles.Role{BaseModel: shared.BaseModel{PublicID: "role_admin"}, Name: "Admin", Slug: "admin"}
	roleTwo := roles.Role{BaseModel: shared.BaseModel{PublicID: "role_user"}, Name: "User", Slug: "user"}
	if err := db.Create(&roleOne).Error; err != nil {
		t.Fatalf("create role one: %v", err)
	}
	if err := db.Create(&roleTwo).Error; err != nil {
		t.Fatalf("create role two: %v", err)
	}
	companyID := uint(1)
	users := []User{
		{BaseModel: shared.BaseModel{PublicID: "u1"}, Name: "One", Email: "one@example.com", PasswordHash: "hash", RoleID: roleOne.ID, CompanyID: &companyID, IsActive: true},
		{BaseModel: shared.BaseModel{PublicID: "u2"}, Name: "Two", Email: "two@example.com", PasswordHash: "hash", RoleID: roleOne.ID, CompanyID: &companyID, IsActive: true},
		{BaseModel: shared.BaseModel{PublicID: "u3"}, Name: "Three", Email: "three@example.com", PasswordHash: "hash", RoleID: roleTwo.ID, CompanyID: &companyID, IsActive: true},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	repo := NewRepository(db)
	count, err := repo.CountByRoleID(roleOne.ID)
	if err != nil {
		t.Fatalf("count users by role id: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2 for role one, got %d", count)
	}
}

func newUserRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&roles.Role{}, &User{}); err != nil {
		t.Fatalf("migrate users test db: %v", err)
	}
	return db
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return parsed
}
