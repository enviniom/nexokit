package app

import (
	"time"

	"log/slog"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/auth"
	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/iam"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/onboarding"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/password"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/enviniom/nexokit/internal/platform/token"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Container holds the dependency graph for all application modules.
type Container struct {
	IAM                *iam.Container
	Companies          *companies.Container
	authContainer      *auth.Container
	Onboarding         *onboarding.Container
	authMW             gin.HandlerFunc
	loginRateLimitMW   gin.HandlerFunc
	refreshRateLimitMW gin.HandlerFunc
}

// NewContainer creates a new Container with the given dependencies.
// Config, DB, Logger and Cache are passed as constructor parameters but are NOT
// stored as public fields; they are only used during module wiring.
func NewContainer(cfg *config.Config, db *gorm.DB, log *slog.Logger, cache cache.Cache, limiter middleware.Limiter) *Container {
	_ = log
	companiesContainer := companies.NewContainer(db)
	iamContainer := iam.NewContainer(db, cache, log)
	passwordManager := password.Manager{}

	tokenManager := token.NewManager(cfg.Auth.PASETOKey, time.Duration(cfg.Auth.AccessTTLMinutes)*time.Minute)
	authContainer := auth.NewContainer(db, passwordManager, tokenManager, time.Duration(cfg.Auth.RefreshTTLDays)*24*time.Hour)
	authMW := middleware.Auth(tokenManager, userLookup{resolver: iamContainer.AuthUserResolver})
	window := time.Duration(cfg.RateLimit.WindowSeconds) * time.Second
	loginRateLimitMW := middleware.RateLimitMiddleware(limiter, cfg.RateLimit.Enabled, "login", cfg.RateLimit.LoginRPM, window)
	refreshRateLimitMW := middleware.RateLimitMiddleware(limiter, cfg.RateLimit.Enabled, "refresh", cfg.RateLimit.RefreshRPM, window)

	onboardingContainer := onboarding.NewContainer(db, onboarding.Config{PasswordHasher: passwordManager, PlatformDomain: cfg.App.PlatformDomain})

	return &Container{
		IAM:                iamContainer,
		Companies:          companiesContainer,
		authContainer:      authContainer,
		Onboarding:         onboardingContainer,
		authMW:             authMW,
		loginRateLimitMW:   loginRateLimitMW,
		refreshRateLimitMW: refreshRateLimitMW,
	}
}

// RegisterModules mounts all business module routes onto the v1 router group.
func (c *Container) RegisterModules(v1 *gin.RouterGroup) {
	auth.Register(v1, c.authContainer, c.authMW, c.loginRateLimitMW, c.refreshRateLimitMW)
	globalProtected := v1.Group("")
	globalProtected.Use(c.authMW, middleware.AllowRootGlobalScope(c.Companies.Resolver()))
	companies.Register(globalProtected, c.Companies, middleware.RequirePermission, middleware.RequireRole)
	tenantProtected := v1.Group("")
	tenantProtected.Use(c.authMW, middleware.RequireTenantScope(c.Companies.Resolver()))
	iam.Register(globalProtected, c.IAM, tenantProtected, middleware.RequirePermission, middleware.RequireRole)
	onboarding.Register(globalProtected, c.Onboarding, middleware.RequireRole)
}

type userLookup struct {
	resolver iamcore.AuthUserResolver
}

type roleResolverAdapter struct {
	resolver iamcore.RoleBySlugResolver
}

func (r roleResolverAdapter) GetBySlug(slug string) (*iamcore.IAMRole, error) {
	return r.resolver.ResolveRoleBySlug(slug)
}

func (l userLookup) GetAuthUser(publicID string) (*authctx.User, error) {
	return l.resolver.ResolveAuthUser(publicID)
}

// SyncPermissions delegates synchronization of registered permissions to IAM service.
func (c *Container) SyncPermissions() error {
	slugs := platformPerms.ListRegistered()
	return c.IAM.Syncer.SyncPermissions(slugs)
}
