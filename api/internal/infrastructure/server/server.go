package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hurtki/github-banners/api/internal/config"
	log "github.com/hurtki/github-banners/api/internal/logger"
	"github.com/rs/cors"
)

type Server struct {
	httpServer *http.Server
	logger     log.Logger
}

func New(cfg *config.Config, handler http.Handler, logger log.Logger) *Server {
	handler = cors.New(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	}).Handler(handler)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Port),
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger.With("service", "http-server"),
	}
}

func (s *Server) Start() {
	go s.run()
}

func (s *Server) run() {
	s.logger.Info("Server is ready to handle requests", "port", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("Could not listen on port", "port", s.httpServer.Addr, "error", err)
	}

	s.logger.Info("Server stopped")
}

func (s *Server) Close(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
