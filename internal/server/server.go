package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/gin-gonic/gin"
)

// Server wraps an http.Server with graceful shutdown capabilities.
type Server struct {
	httpServer *http.Server
}

// New creates a new Server instance.
func New(cfg *config.Config, router *gin.Engine) *Server {
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start begins listening for incoming connections.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server with the given timeout context.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
