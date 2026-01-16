package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hurtki/github-banners/api/internal/config"
	log "github.com/hurtki/github-banners/api/internal/logger"
)

type Server struct {
	httpServer *http.Server
	logger     log.Logger
}

func New(cfg *config.Config, handler http.Handler, logger log.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Port),
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		s.logger.Info("Shutting down server")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Could not gracefully shutdown server", "error", err)
		}

		close(done)
	}()

	s.logger.Info("Server is ready to handle requests", "port", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("Could not listen on port", "port", s.httpServer.Addr, "error", err)
		return err
	}

	<-done
	s.logger.Info("Server stopped")
	return nil
}
