package main

import (
	"context"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	user_stats_worker "github.com/hurtki/github-banners/api/internal/app/user_stats"
	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/hurtki/github-banners/api/internal/handlers"
	infraDB "github.com/hurtki/github-banners/api/internal/infrastructure/db"
	infraGithub "github.com/hurtki/github-banners/api/internal/infrastructure/github"
	"github.com/hurtki/github-banners/api/internal/infrastructure/server"
	log "github.com/hurtki/github-banners/api/internal/logger"
	"github.com/hurtki/github-banners/api/internal/repo/github_user_data"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger := log.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting github banners API service")

	psgrConf, err := config.LoadPostgres()
	if err != nil {
		logger.Error("Can't load postgres database config", "err", err.Error())
		os.Exit(1)

	}

	// Create in-memory cache
	memoryCache := cache.NewCache(cfg.CacheTTL)

	// Create service configuration
	serviceConfig := &domain.ServiceConfig{
		CacheTTL:       cfg.CacheTTL,
		RequestTimeout: cfg.RequestTimeout,
	}
	// Create GitHub fetcher (infrastructure layer)
	githubFetcher := infraGithub.NewFetcher(cfg.GithubToken, serviceConfig)

	db, err := infraDB.NewDB(psgrConf, logger)
	if err != nil {
		logger.Error("can't intialize database, exiting", "err", err.Error())
		os.Exit(1)
	}
	repo := github_user_data.NewGithubDataPsgrRepo(db, logger)

	// Create stats service (domain service with cache)
	statsService := userstats.NewUserStatsService(repo, githubFetcher, memoryCache)

	// TODO add configurating of worker to app config from env variables
	statsWorker := user_stats_worker.NewStatsWorker(statsService.RefreshAll, time.Hour, logger, userstats.WorkerConfig{BatchSize: 1, Concurrency: 1, CacheTTL: time.Hour})
	statsWorker.Start(context.TODO())

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
