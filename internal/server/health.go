package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type dbPinger interface {
	PingContext(context.Context) error
}

type cacheGetter interface {
	Get(context.Context, string) ([]byte, error)
}

type HealthDeps struct {
	DB           dbPinger
	Cache        cacheGetter
	CacheEnabled bool
}

type LiveResponse struct {
	Status string `json:"status"`
}

type DepStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type ReadyResponse struct {
	Dependencies []DepStatus `json:"dependencies"`
}

func liveHandler(c *gin.Context) {
	c.JSON(http.StatusOK, LiveResponse{Status: "alive"})
}

func readyHandler(deps HealthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		statuses := make([]DepStatus, 0, 2)
		healthy := true

		dbStatus := DepStatus{Name: "database", Status: "healthy"}
		if deps.DB == nil {
			dbStatus.Status = "unhealthy"
			dbStatus.Error = "database dependency is not configured"
			healthy = false
		} else if err := deps.DB.PingContext(ctx); err != nil {
			dbStatus.Status = "unhealthy"
			dbStatus.Error = err.Error()
			healthy = false
		}
		statuses = append(statuses, dbStatus)

		cacheStatus := DepStatus{Name: "cache", Status: "healthy"}
		if deps.CacheEnabled {
			if deps.Cache == nil {
				cacheStatus.Status = "unhealthy"
				cacheStatus.Error = "cache dependency is not configured"
				healthy = false
			} else if _, err := deps.Cache.Get(ctx, "health:probe"); err != nil && !errors.Is(err, context.Canceled) {
				cacheStatus.Status = "unhealthy"
				cacheStatus.Error = err.Error()
				healthy = false
			}
		}
		statuses = append(statuses, cacheStatus)

		statusCode := http.StatusOK
		if !healthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReadyResponse{Dependencies: statuses})
	}
}

func healthHandler(c *gin.Context) {
	response.Success(c, messages.MsgHealthy, map[string]string{"status": "ok"})
}
