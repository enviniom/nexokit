package app

import (
	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
)

// Container holds the dependency graph for all application modules.
// It is intentionally minimal in change-01 and will grow as modules are implemented.
type Container struct{}

// NewContainer creates a new Container with the given dependencies.
// Config, DB, Logger and Cache are passed as constructor parameters but are NOT
// stored as public fields; they are only used during module wiring.
func NewContainer(cfg *config.Config, db *gorm.DB, log *slog.Logger, cache cache.Cache) *Container {
	// fields are intentionally not stored; only used during wiring
	return &Container{}
}

// RegisterModules mounts all business module routes onto the v1 router group.
// In change-01, no modules are registered yet.
func (c *Container) RegisterModules(v1 *gin.RouterGroup) {
	// TODO: register auth, users, roles, companies modules in future changes
}
