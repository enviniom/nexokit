package app

import (
	"time"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/modules/auth"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/password"
	"github.com/enviniom/nexokit/internal/platform/token"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
)

// Container holds the dependency graph for all application modules.
type Container struct {
	rolesHandler *roles.Handler
	usersHandler *users.Handler
	authHandler  *auth.Handler
	authMW       gin.HandlerFunc
}

// NewContainer creates a new Container with the given dependencies.
// Config, DB, Logger and Cache are passed as constructor parameters but are NOT
// stored as public fields; they are only used during module wiring.
func NewContainer(cfg *config.Config, db *gorm.DB, log *slog.Logger, cache cache.Cache) *Container {
	_ = log
	_ = cache

	usersRepo := users.NewRepository(db)
	permissionsRepo := permissions.NewRepository(db)

	rolesRepo := roles.NewRepository(db)
	rolesService := roles.NewService(rolesRepo, roles.WithPermissionCatalog(permissionsRepo), roles.WithRoleMembers(usersRepo), roles.WithCache(cache))
	rolesHandler := roles.NewHandler(rolesService)

	passwordManager := password.Manager{}
	usersService := users.NewService(usersRepo, passwordManager, rolesRepo)
	usersHandler := users.NewHandler(usersService, authctx.PublicIDFromGin)

	tokenManager := token.NewManager(cfg.Auth.PASETOKey, time.Duration(cfg.Auth.AccessTTLMinutes)*time.Minute)
	refreshRepo := auth.NewRefreshRepository(db)
	authService := auth.NewService(usersRepo, passwordManager, tokenManager, tokenManager, refreshRepo, time.Duration(cfg.Auth.RefreshTTLDays)*24*time.Hour)
	authHandler := auth.NewHandler(authService)
	authMW := middleware.Auth(tokenManager, userLookup{repo: usersRepo})

	return &Container{rolesHandler: rolesHandler, usersHandler: usersHandler, authHandler: authHandler, authMW: authMW}
}

// RegisterModules mounts all business module routes onto the v1 router group.
func (c *Container) RegisterModules(v1 *gin.RouterGroup) {
	auth.Register(v1, c.authHandler, c.authMW)
	protected := v1.Group("")
	protected.Use(c.authMW)
	roles.Register(protected, c.rolesHandler)
	users.Register(protected, c.usersHandler)
}

type userLookup struct {
	repo users.Repository
}

func (l userLookup) GetAuthUser(publicID string) (*authctx.User, error) {
	user, err := l.repo.GetByPublicID(publicID)
	if err != nil {
		return nil, err
	}
	return &authctx.User{
		ID:        user.ID,
		PublicID:  user.PublicID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role.Name,
		RoleID:    user.RoleID,
		CompanyID: user.CompanyID,
		IsActive:  user.IsActive,
	}, nil
}
