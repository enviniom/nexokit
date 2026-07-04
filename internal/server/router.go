package server

import (
	"io"
	"log/slog"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/gin-gonic/gin"
)

// NewRouter creates a configured gin.Engine with middleware and route registration.
func NewRouter(cfg *config.Config, log, errorLog *slog.Logger, ginWriter io.Writer, healthDeps HealthDeps, registerModules func(*gin.RouterGroup)) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.IsTest() {
		gin.SetMode(gin.TestMode)
	}

	r := gin.New()

	// Middleware chain: RequestID → DebugErrors → GinLogger → Logger → ErrorLogger → Recovery → CORS
	// DebugErrors stores the config-derived debug flag so response.HandleError
	// does not depend on the global gin mode. ErrorLogger is registered before
	// Recovery because Gin unwinds middleware in reverse order; this lets
	// Recovery push panic errors into c.Errors before ErrorLogger's post-c.Next()
	// body runs and owns the single log line.
	r.Use(middleware.RequestID())
	r.Use(middleware.DebugErrors(cfg.ExposeDebugErrors()))
	r.Use(gin.LoggerWithWriter(ginWriter))
	r.Use(middleware.Logger(log))
	r.Use(middleware.ErrorLogger(errorLog))
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(cfg))

	// Health checks (unversioned)
	r.GET("/health", healthHandler)
	r.GET("/health/live", liveHandler)
	r.GET("/health/ready", readyHandler(healthDeps))

	// API v1 group
	v1 := r.Group("/api/v1")
	if registerModules != nil {
		registerModules(v1)
	}

	return r
}
