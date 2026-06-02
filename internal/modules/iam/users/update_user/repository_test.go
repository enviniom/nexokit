package update_user

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func seedSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&core.IAMRole{}, &core.IAMUser{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
}

func seedRole(t *testing.T, db *gorm.DB, publicID, name, slug string) core.IAMRole {
	t.Helper()
	role := core.IAMRole{BaseModel: shared.BaseModel{PublicID: publicID}, Name: name, Slug: slug}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return role
}

func seedUser(t *testing.T, db *gorm.DB, publicID, name, email string, roleID uint, companyID *uint) core.IAMUser {
	t.Helper()
	user := core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: publicID},
		Name:         name,
		Email:        email,
		PasswordHash: "hash",
		RoleID:       roleID,
		CompanyID:    companyID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}

func TestRepositoryGetByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")
	companyID := uint(42)
	seedUser(t, db, "user-1", "Alice", "alice@example.com", role.ID, &companyID)

	repo := NewRepository(db)

	// NOTE: GetByPublicID delegates to queries.GetUserByPublicID.
	// Full query behavior coverage lives in queries/get_user_by_public_id_test.go.
	user, err := repo.GetByPublicID(tenant.NewRoot(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.PublicID != "user-1" {
		t.Fatalf("expected public id user-1, got %s", user.PublicID)
	}
	if user.Role.Slug != "admin" {
		t.Fatalf("expected role slug admin, got %s", user.Role.Slug)
	}

	_, err = repo.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepositoryGetRoleBySlug(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	seedRole(t, db, "role-root", "Root", "root")

	repo := NewRepository(db)

	role, err := repo.GetRoleBySlug("root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Slug != "root" {
		t.Fatalf("expected slug root, got %s", role.Slug)
	}

	_, err = repo.GetRoleBySlug("missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepositoryUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")
	user := seedUser(t, db, "user-1", "Alice", "alice@example.com", role.ID, nil)

	repo := NewRepository(db)

	user.Name = "Updated Alice"
	user.Email = "updated@example.com"
	if err := repo.Update(&user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var reloaded core.IAMUser
	if err := db.First(&reloaded, "public_id = ?", "user-1").Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Name != "Updated Alice" {
		t.Fatalf("expected name Updated Alice, got %s", reloaded.Name)
	}
	if reloaded.Email != "updated@example.com" {
		t.Fatalf("expected email updated@example.com, got %s", reloaded.Email)
	}
}

func TestRepositoryUpdateDuplicateEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")
	seedUser(t, db, "user-1", "Alice", "alice@example.com", role.ID, nil)
	user2 := seedUser(t, db, "user-2", "Bob", "bob@example.com", role.ID, nil)

	repo := NewRepository(db)

	user2.Email = "alice@example.com"
	err = repo.Update(&user2)
	if !errors.Is(err, core.ErrUserEmailAlreadyExists) {
		t.Fatalf("expected ErrUserEmailAlreadyExists, got %v", err)
	}
}

func TestRepositoryReload(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")
	companyID := uint(42)
	seedUser(t, db, "user-1", "Alice", "alice@example.com", role.ID, &companyID)

	repo := NewRepository(db)

	// NOTE: Reload delegates to queries.GetUserByPublicID.
	// Full query behavior coverage lives in queries/get_user_by_public_id_test.go.
	resp, err := repo.Reload(tenant.NewRoot(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PublicID != "user-1" {
		t.Fatalf("expected public id user-1, got %s", resp.PublicID)
	}
	if resp.RoleName != "Admin" {
		t.Fatalf("expected role name Admin, got %s", resp.RoleName)
	}
	if resp.CompanyID == nil || *resp.CompanyID != 42 {
		t.Fatalf("expected company_id 42, got %v", resp.CompanyID)
	}

	_, err = repo.Reload(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
