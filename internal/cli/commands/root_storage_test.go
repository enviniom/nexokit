package commands

import (
	"testing"

	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := db.AutoMigrate(&roles.Role{}, &users.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestRootStorage_RootExists_NoRootRole(t *testing.T) {
	db := setupTestDB(t)
	storage := newRootStorage(db)

	exists, err := storage.RootExists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected RootExists to be false when no root role exists")
	}
}

func TestRootStorage_RootExists_RoleButNoUser(t *testing.T) {
	db := setupTestDB(t)
	storage := newRootStorage(db)

	rootRole := roles.Role{Name: "root", Slug: roles.RootRoleSlug, IsSystem: true}
	if err := db.Create(&rootRole).Error; err != nil {
		t.Fatalf("failed to create root role: %v", err)
	}

	exists, err := storage.RootExists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected RootExists to be false when root role exists but no user")
	}
}

func TestRootStorage_RootExists_UserExists(t *testing.T) {
	db := setupTestDB(t)
	storage := newRootStorage(db)

	rootRole := roles.Role{Name: "root", Slug: roles.RootRoleSlug, IsSystem: true}
	if err := db.Create(&rootRole).Error; err != nil {
		t.Fatalf("failed to create root role: %v", err)
	}

	user := &users.User{
		BaseModel: shared.BaseModel{PublicID: "usr01"},
		Name:      "Root",
		Email:     "root@example.com",
		RoleID:    rootRole.ID,
		IsActive:  true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	exists, err := storage.RootExists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected RootExists to be true when root user exists")
	}
}

func TestRootStorage_CreateRoot_Success(t *testing.T) {
	db := setupTestDB(t)
	storage := newRootStorage(db)

	rootRole := roles.Role{Name: "root", Slug: roles.RootRoleSlug, IsSystem: true}
	if err := db.Create(&rootRole).Error; err != nil {
		t.Fatalf("failed to create root role: %v", err)
	}

	if err := storage.CreateRoot("Root User", "root@example.com", "hashed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	if err := db.Model(&users.User{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}

	var created users.User
	if err := db.First(&created).Error; err != nil {
		t.Fatalf("failed to fetch created user: %v", err)
	}
	if created.Name != "Root User" {
		t.Errorf("expected name 'Root User', got %s", created.Name)
	}
	if created.Email != "root@example.com" {
		t.Errorf("expected email 'root@example.com', got %s", created.Email)
	}
	if created.PasswordHash != "hashed" {
		t.Errorf("expected password hash 'hashed', got %s", created.PasswordHash)
	}
	if created.RoleID != rootRole.ID {
		t.Errorf("expected role_id %d, got %d", rootRole.ID, created.RoleID)
	}
	if !created.IsActive {
		t.Error("expected user to be active")
	}
	if created.PublicID == "" {
		t.Error("expected user to have a public_id")
	}
}

func TestRootStorage_CreateRoot_MissingRole(t *testing.T) {
	db := setupTestDB(t)
	storage := newRootStorage(db)

	err := storage.CreateRoot("Root", "root@example.com", "hashed")
	if err == nil {
		t.Fatal("expected error when root role is missing")
	}
}
