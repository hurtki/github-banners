package main

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/github"
	"github.com/hurtki/github-banners/api/internal/handlers"
	infraGithub "github.com/hurtki/github-banners/api/internal/infrastructure/github"
	"github.com/hurtki/github-banners/api/internal/infrastructure/server"
	log "github.com/hurtki/github-banners/api/internal/logger"
	"github.com/hurtki/github-banners/api/internal/service"
)

func main() {
	cfg := config.Load()

	logger := log.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting GitHub Stats API")

	memoryCache := cache.NewCache(cfg.CacheTTL)

	serviceConfig := &domain.ServiceConfig{
		CacheTTL:       cfg.CacheTTL,
		RequestTimeout: cfg.RequestTimeout,
	}

	githubFetcher := infraGithub.NewFetcher(cfg.GithubToken, serviceConfig)

	statsService := service.NewStatsService(githubFetcher, memoryCache)

	router := chi.NewRouter()

	// Register GitHub stats handler routes
	githubHandler := github.NewHandler(cfg, logger, statsService)
	githubHandler.RegisterRoutes(router)

	// Register banners handler routes
	bannersHandler := handlers.NewBannersHandler(logger, statsService)
	router.Get("/banners/preview", bannersHandler.Preview)
	router.Post("/banners", bannersHandler.Create)

	srv := server.New(cfg, router, logger)
	if err := srv.Start(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
