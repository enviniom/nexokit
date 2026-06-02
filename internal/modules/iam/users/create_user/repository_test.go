package create_user

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

func TestRepositoryExistsByEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")

	user := core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: ExistsByEmail delegates to queries.GetUserByEmail.
	// Full query behavior coverage lives in queries/get_user_by_email_test.go.
	exists, err := repo.ExistsByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatalf("expected email to exist")
	}

	exists, err = repo.ExistsByEmail("nobody@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("expected email to not exist")
	}
}

func TestRepositoryCreate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")

	repo := NewRepository(db)

	user := &core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID == 0 {
		t.Fatalf("expected persisted ID")
	}
}

func TestRepositoryCreateDuplicateEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")

	repo := NewRepository(db)

	first := &core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := repo.Create(first); err != nil {
		t.Fatalf("first create: %v", err)
	}

	duplicate := &core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-2"},
		Name:         "Alice Clone",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		IsActive:     true,
	}
	err = repo.Create(duplicate)
	if !errors.Is(err, core.ErrUserEmailAlreadyExists) {
		t.Fatalf("expected ErrUserEmailAlreadyExists, got %v", err)
	}
}

func TestRepositoryGetByPublicID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	seedSchema(t, db)
	role := seedRole(t, db, "role-1", "Admin", "admin")

	companyID := uint(42)
	user := core.IAMUser{
		BaseModel:    shared.BaseModel{PublicID: "user-1"},
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		RoleID:       role.ID,
		CompanyID:    &companyID,
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repo := NewRepository(db)

	// NOTE: GetByPublicID delegates to queries.GetUserByPublicID.
	// Full query behavior coverage lives in queries/get_user_by_public_id_test.go.
	item, err := repo.GetByPublicID(tenant.NewRoot(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.PublicID != "user-1" {
		t.Fatalf("expected public id user-1, got %s", item.PublicID)
	}
	if item.RoleName != "Admin" {
		t.Fatalf("expected role name Admin, got %s", item.RoleName)
	}
	if item.CompanyID == nil || *item.CompanyID != 42 {
		t.Fatalf("expected company_id 42, got %v", item.CompanyID)
	}

	_, err = repo.GetByPublicID(tenant.NewRoot(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
