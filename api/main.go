package main

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/domain/services"
	"github.com/hurtki/github-banners/api/internal/handlers"
	infraGithub "github.com/hurtki/github-banners/api/internal/infrastructure/github"
	"github.com/hurtki/github-banners/api/internal/infrastructure/server"
	log "github.com/hurtki/github-banners/api/internal/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger := log.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting GitHub Stats API")

	// Create in-memory cache
	memoryCache := cache.NewCache(cfg.CacheTTL)

	// Create service configuration
	serviceConfig := &domain.ServiceConfig{
		CacheTTL:       cfg.CacheTTL,
		RequestTimeout: cfg.RequestTimeout,
	}
	// Create GitHub fetcher (infrastructure layer)
	githubFetcher := infraGithub.NewFetcher(cfg.GithubToken, serviceConfig)

	// Create stats service (domain service with cache)
	statsService := services.NewStatsService(githubFetcher, memoryCache, cfg.CacheTTL)

	router := chi.NewRouter()

	bannersHandler := handlers.NewBannersHandler(logger, statsService)

	router.Get("/banners/preview", bannersHandler.Preview)
	router.Post("/banners", bannersHandler.Create)

	// Create and start HTTP server
	srv := server.New(cfg, router, logger)
	if err := srv.Start(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
