package main

import (
	"os"

	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/github"
	infraGithub "github.com/hurtki/github-banners/api/internal/infrastructure/github"
	"github.com/hurtki/github-banners/api/internal/infrastructure/server"
	log "github.com/hurtki/github-banners/api/internal/logger"
	"github.com/hurtki/github-banners/api/internal/service"
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
	statsService := service.NewStatsService(githubFetcher, memoryCache)

	// Create HTTP handler
	handler := github.NewHandler(cfg, logger, statsService)

	// Create and start HTTP server
	srv := server.New(cfg, handler, logger)
	if err := srv.Start(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
