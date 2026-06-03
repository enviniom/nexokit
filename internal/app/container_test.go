package app

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/iam"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegisterModules_MountsIAMEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	c := NewContainer(minimalConfig(), db, slog.Default(), cache.NewNoop(), middleware.NewNoopLimiter())
	r := gin.New()
	v1 := r.Group("/api/v1")
	c.RegisterModules(v1)

	routeSet := map[string]struct{}{}
	for _, route := range r.Routes() {
		routeSet[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"GET /api/v1/users",
		"POST /api/v1/users",
		"GET /api/v1/users/:id",
		"PUT /api/v1/users/:id",
		"DELETE /api/v1/users/:id",
		"PATCH /api/v1/users/:id/password",
		"PATCH /api/v1/users/:id/role",
		"PATCH /api/v1/users/:id/status",
		"GET /api/v1/roles",
		"GET /api/v1/roles/select",
		"GET /api/v1/roles/:id",
		"POST /api/v1/roles",
		"PUT /api/v1/roles/:id",
		"DELETE /api/v1/roles/:id",
		"GET /api/v1/roles/:id/permissions",
		"PUT /api/v1/roles/:id/permissions",
		"GET /api/v1/permissions",
		"GET /api/v1/permissions/:id",
		"PUT /api/v1/permissions/:id",
	}

	for _, key := range expected {
		if _, ok := routeSet[key]; !ok {
			t.Fatalf("missing IAM route %q", key)
		}
	}
}

func TestUserLookup_DelegatesToIAMResolver(t *testing.T) {
	expected := &authctx.User{PublicID: "u-1", IsActive: true}
	lookup := userLookup{resolver: fakeAuthUserResolver{user: expected}}

	got, err := lookup.GetAuthUser("u-1")
	if err != nil {
		t.Fatalf("GetAuthUser returned error: %v", err)
	}
	if got != expected {
		t.Fatalf("expected delegated user pointer")
	}
}

func TestSyncPermissions_DelegatesToIAMSyncer(t *testing.T) {
	slug := "iam.test.sync"
	platformPerms.Register(slug)

	syncer := &fakePermissionSyncer{}
	c := &Container{IAM: &iam.Container{Syncer: syncer}}

	if err := c.SyncPermissions(); err != nil {
		t.Fatalf("SyncPermissions returned error: %v", err)
	}
	if !slices.Contains(syncer.slugs, slug) {
		t.Fatalf("expected syncer to receive %q, got %v", slug, syncer.slugs)
	}
}

func minimalConfig() *config.Config {
	return &config.Config{
		App:       config.AppConfig{PlatformDomain: "example.test"},
		RateLimit: config.RateLimitConfig{Enabled: false, WindowSeconds: 60, LoginRPM: 5, RefreshRPM: 10},
		Auth:      config.AuthConfig{PASETOKey: "test-secret", AccessTTLMinutes: 15, RefreshTTLDays: 7},
	}
}

type fakeAuthUserResolver struct{ user *authctx.User }

func (f fakeAuthUserResolver) ResolveAuthUser(string) (*authctx.User, error) { return f.user, nil }

type fakeRoleBySlugResolver struct{ role *iamcore.IAMRole }

func (f fakeRoleBySlugResolver) ResolveRoleBySlug(string) (*iamcore.IAMRole, error) {
	return f.role, nil
}

type fakePermissionSyncer struct{ slugs []string }

func (f *fakePermissionSyncer) SyncPermissions(slugs []string) error {
	f.slugs = append([]string(nil), slugs...)
	return nil
}
