package app

import (
	"time"

	"log/slog"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/auth"
	"github.com/enviniom/nexokit/internal/modules/companies"
	"github.com/enviniom/nexokit/internal/modules/onboarding"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/password"
	platformPerms "github.com/enviniom/nexokit/internal/platform/permissions"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/platform/token"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Container holds the dependency graph for all application modules.
type Container struct {
	rolesHandler         *roles.Handler
	usersHandler         *users.Handler
	Companies            *companies.Container
	permissionsContainer *permissions.Container
	authContainer        *auth.Container
	Onboarding           *onboarding.Container
	authMW               gin.HandlerFunc
	authzMW              gin.HandlerFunc
	loginRateLimitMW     gin.HandlerFunc
	refreshRateLimitMW   gin.HandlerFunc
}

// NewContainer creates a new Container with the given dependencies.
// Config, DB, Logger and Cache are passed as constructor parameters but are NOT
// stored as public fields; they are only used during module wiring.
func NewContainer(cfg *config.Config, db *gorm.DB, log *slog.Logger, cache cache.Cache, limiter middleware.Limiter) *Container {
	_ = log

	usersRepo := users.NewRepository(db)
	companiesContainer := companies.NewContainer(db)

	permissionsContainer := permissions.NewContainer(db, cache, log)

	rolesRepo := roles.NewRepository(db)
	rolesService := roles.NewService(rolesRepo, roles.WithPermissionCatalog(permissionsContainer.Catalog), roles.WithRoleMembers(usersRepo), roles.WithCache(cache))
	rolesHandler := roles.NewHandler(rolesService)

	passwordManager := password.Manager{}
	usersService := users.NewService(usersRepo, passwordManager, roleResolverAdapter{repo: rolesRepo})
	usersHandler := users.NewHandler(usersService, authctx.PublicIDFromGin)

	tokenManager := token.NewManager(cfg.Auth.PASETOKey, time.Duration(cfg.Auth.AccessTTLMinutes)*time.Minute)
	authContainer := auth.NewContainer(db, passwordManager, tokenManager, time.Duration(cfg.Auth.RefreshTTLDays)*24*time.Hour)
	authMW := middleware.Auth(tokenManager, userLookup{repo: usersRepo})
	authzMW := middleware.AttachPermissions(permissionsContainer.Resolver)
	window := time.Duration(cfg.RateLimit.WindowSeconds) * time.Second
	loginRateLimitMW := middleware.RateLimitMiddleware(limiter, cfg.RateLimit.Enabled, "login", cfg.RateLimit.LoginRPM, window)
	refreshRateLimitMW := middleware.RateLimitMiddleware(limiter, cfg.RateLimit.Enabled, "refresh", cfg.RateLimit.RefreshRPM, window)

	onboardingContainer := onboarding.NewContainer(db, onboarding.Config{PasswordHasher: passwordManager, PlatformDomain: cfg.App.PlatformDomain})

	return &Container{
		rolesHandler:         rolesHandler,
		usersHandler:         usersHandler,
		Companies:            companiesContainer,
		permissionsContainer: permissionsContainer,
		authContainer:        authContainer,
		Onboarding:           onboardingContainer,
		authMW:               authMW,
		authzMW:              authzMW,
		loginRateLimitMW:     loginRateLimitMW,
		refreshRateLimitMW:   refreshRateLimitMW,
	}
}

// RegisterModules mounts all business module routes onto the v1 router group.
func (c *Container) RegisterModules(v1 *gin.RouterGroup) {
	auth.Register(v1, c.authContainer, c.authMW, c.authzMW, c.loginRateLimitMW, c.refreshRateLimitMW)
	globalProtected := v1.Group("")
	globalProtected.Use(c.authMW, middleware.AllowRootGlobalScope(c.Companies.Resolver()), c.authzMW)
	companies.Register(globalProtected, c.Companies, middleware.RequirePermission, middleware.RequireRole)
	// Roles and permissions are system catalog modules, so root may administer
	// them globally while non-root requests remain scoped to their company.
	roles.Register(globalProtected, c.rolesHandler, middleware.RequirePermission)
	permissions.Register(globalProtected, c.permissionsContainer, middleware.RequirePermission)
	onboarding.Register(globalProtected, c.Onboarding, middleware.RequireRole)

	tenantProtected := v1.Group("")
	tenantProtected.Use(c.authMW, middleware.RequireTenantScope(c.Companies.Resolver()), c.authzMW)
	users.Register(tenantProtected, c.usersHandler, middleware.RequirePermission)
}

type userLookup struct {
	repo users.Repository
}

type roleResolverAdapter struct {
	repo roles.Repository
}

func (r roleResolverAdapter) GetBySlug(slug string) (*roles.Role, error) {
	return r.repo.GetBySlug(tenant.NewRoot(), slug)
}

func (l userLookup) GetAuthUser(publicID string) (*authctx.User, error) {
	user, err := l.repo.GetAuthUser(publicID)
	if err != nil {
		return nil, err
	}
	return &authctx.User{
		ID:        user.ID,
		PublicID:  user.PublicID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role.Name,
		RoleSlug:  user.Role.Slug,
		RoleID:    user.RoleID,
		CompanyID: user.CompanyID,
		IsRoot:    user.IsRoot(),
		IsActive:  user.IsActive,
	}, nil
}

// SyncPermissions delegates synchronization of registered permissions to permissionsService.
func (c *Container) SyncPermissions() error {
	slugs := platformPerms.ListRegistered()
	return c.permissionsContainer.Syncer.SyncPermissions(slugs)
}
