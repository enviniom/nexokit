package server

import (
	"io"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
	"log/slog"
)

// NewRouter creates a configured gin.Engine with middleware and route registration.
func NewRouter(cfg *config.Config, log *slog.Logger, ginWriter io.Writer, registerModules func(*gin.RouterGroup)) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.IsTest() {
		gin.SetMode(gin.TestMode)
	}

	r := gin.New()

	// Middleware chain: RequestID → GinLogger → Logger → Recovery → CORS
	r.Use(middleware.RequestID())
	r.Use(gin.LoggerWithWriter(ginWriter))
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS(cfg))

	// Health check (unversioned)
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, messages.MsgHealthy, map[string]string{"status": "ok"})
	})

	// API v1 group
	v1 := r.Group("/api/v1")
	if registerModules != nil {
		registerModules(v1)
	}

	return r
}
