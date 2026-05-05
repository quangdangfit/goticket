// Package server builds the HTTP layer: gin engine, global middleware,
// route registration. Domain handlers are mounted by main.go via
// RegisterDomain.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quangdangfit/goticket/config"
	"github.com/quangdangfit/goticket/internal/server/middleware"
	"github.com/quangdangfit/goticket/pkg/logger"
)

// Server wraps a *http.Server and the gin engine.
type Server struct {
	cfg    config.AppConfig
	engine *gin.Engine
	http   *http.Server
}

// New builds a Server with global middleware and ops endpoints.
func New(cfg config.AppConfig) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recover(), middleware.Metrics(), gin.Logger())
	registerOps(r)
	return &Server{
		cfg:    cfg,
		engine: r,
		http: &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

// APIGroup returns the root /api/v1 router group; domain packages register
// their routes under this group.
func (s *Server) APIGroup() *gin.RouterGroup { return s.engine.Group("/api/v1") }

// WebhookGroup returns the /webhook prefix for provider callbacks.
func (s *Server) WebhookGroup() *gin.RouterGroup { return s.engine.Group("/webhook") }

// Run starts the HTTP server and blocks until ctx is cancelled, then
// gracefully shuts down within ShutdownTimeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		logger.FromContext(ctx).Info("http listen", "addr", s.cfg.HTTPAddr)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		return s.http.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}
