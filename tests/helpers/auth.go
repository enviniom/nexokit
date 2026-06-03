package helpers

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/companies"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/token"
	"gorm.io/gorm"
)

type UserOptions struct {
	PublicID  string
	Email     string
	Name      string
	RoleSlug  string
	CompanyID *uint
	IsActive  bool
}

type Actor struct {
	User iamcore.IAMUser
	Role iamcore.IAMRole
}

func SeedUser(t *testing.T, db *gorm.DB, opts UserOptions) iamcore.IAMUser {
	t.Helper()

	roleSlug := opts.RoleSlug
	if roleSlug == "" {
		roleSlug = iamcore.UserRoleSlug
	}

	role := SeedRole(t, db, roleSlug)
	user := iamcore.IAMUser{
		Name:         opts.Name,
		Email:        opts.Email,
		PasswordHash: "hash",
		RoleID:       role.ID,
		CompanyID:    opts.CompanyID,
		IsActive:     true,
	}
	if user.Name == "" {
		user.Name = "Test User"
	}
	if user.Email == "" {
		user.Email = "test@example.com"
	}
	if opts.PublicID != "" {
		user.PublicID = opts.PublicID
	}
	if !opts.IsActive {
		user.IsActive = false
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed auth user: %v", err)
	}

	return user
}

func CreateTestToken(t *testing.T, db *gorm.DB, user iamcore.IAMUser) string {
	t.Helper()

	var withRole iamcore.IAMUser
	if err := db.Preload("Role").First(&withRole, user.ID).Error; err != nil {
		t.Fatalf("load user role: %v", err)
	}

	manager := token.NewManager("nexokit-test-secret", time.Hour)
	tokenStr, err := manager.IssueAccess(withRole.PublicID, withRole.Role.Slug, withRole.CompanyID)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	return tokenStr
}

func AuthenticatedRequest(t *testing.T, method, path string, body io.Reader, actor Actor) *http.Request {
	t.Helper()

	req := httptestRequest(method, path, body)
	if actor.User.PublicID == "" {
		t.Fatalf("actor user public id is required")
	}

	manager := token.NewManager("nexokit-test-secret", time.Hour)
	tokenStr, err := manager.IssueAccess(actor.User.PublicID, actor.Role.Slug, actor.User.CompanyID)
	if err != nil {
		t.Fatalf("issue request access token: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+tokenStr)
	return req
}

func httptestRequest(method, path string, body io.Reader) *http.Request {
	if body == nil {
		body = http.NoBody
	}
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		panic(err)
	}
	return req
}

func SeedAuthActor(t *testing.T, db *gorm.DB, opts UserOptions) Actor {
	t.Helper()
	if opts.RoleSlug == "" {
		opts.RoleSlug = iamcore.AdminRoleSlug
	}
	if opts.CompanyID != nil {
		company := companies.Company{}
		if err := db.First(&company, *opts.CompanyID).Error; err != nil {
			t.Fatalf("company for actor not found: %v", err)
		}
	}
	user := SeedUser(t, db, opts)
	role := iamcore.IAMRole{}
	if err := db.First(&role, user.RoleID).Error; err != nil {
		t.Fatalf("load actor role: %v", err)
	}
	return Actor{User: user, Role: role}
}
