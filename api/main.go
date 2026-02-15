package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	user_stats_worker "github.com/hurtki/github-banners/api/internal/app/user_stats"
	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/domain/preview"
	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/hurtki/github-banners/api/internal/handlers"
	infraDB "github.com/hurtki/github-banners/api/internal/infrastructure/db"
	infraGithub "github.com/hurtki/github-banners/api/internal/infrastructure/github"
	http_auth "github.com/hurtki/github-banners/api/internal/infrastructure/httpauth"
	"github.com/hurtki/github-banners/api/internal/infrastructure/renderer"
	renderer_http "github.com/hurtki/github-banners/api/internal/infrastructure/renderer/http"
	"github.com/hurtki/github-banners/api/internal/infrastructure/server"
	"github.com/hurtki/github-banners/api/internal/infrastructure/storage"
	log "github.com/hurtki/github-banners/api/internal/logger"
	"github.com/hurtki/github-banners/api/internal/migrations"
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
	statsCache := cache.NewStatsMemoryCache(cfg.CacheTTL)

	// Create service configuration
	serviceConfig := &domain.ServiceConfig{
		CacheTTL:       cfg.CacheTTL,
		RequestTimeout: cfg.RequestTimeout,
	}
	// Create GitHub fetcher (infrastructure layer)
	githubFetcher := infraGithub.NewFetcher(cfg.GithubTokens, serviceConfig, logger)

	db, err := infraDB.NewDB(psgrConf, logger)
	if err != nil {
		logger.Error("can't initialize database, existing", "err", err.Error())
		os.Exit(1)
	}
	if err := migrations.RunMigrations(db); err != nil {
		logger.Error("failed to run migrations", "err", err.Error())
		os.Exit(1)
	}
	repo := github_user_data.NewGithubDataPsgrRepo(db, logger)

	// Create stats service (domain service with cache)
	statsService := userstats.NewUserStatsService(repo, githubFetcher, statsCache)

	// TODO add configurating of worker to app config from env variables
	statsWorker := user_stats_worker.NewStatsWorker(statsService.RefreshAll, time.Hour, logger, userstats.WorkerConfig{BatchSize: 1, Concurrency: 1, CacheTTL: time.Hour})
	go statsWorker.Start(context.TODO())

	router := chi.NewRouter()

	// renderer infra intialization
	rendererAuthRT := http_auth.NewAuthHTTPRoundTripper("api", http_auth.NewHMACSigner([]byte(cfg.ServicesSecret)), time.Now)
	rendererHTTPClient := renderer_http.NewRendererHTTPClient(rendererAuthRT)
	renderer := renderer.NewRenderer(rendererHTTPClient, logger, cfg.RendererBaseURL)

	// storage infra initialization
	storageAuthRT := http_auth.NewAuthHTTPRoundTripper(
		"api",
		http_auth.NewHMACSigner([]byte(cfg.ServicesSecret)),
		time.Now,
	)

	storageHTTPClient := &http.Client{
		Transport: storageAuthRT,
	}

	storageClient := storage.NewClient(
    cfg.StorageBaseURL,
    storageHTTPClient,
    logger,
	)

	previewUsecase := preview.NewPreviewUsecase(statsService, preview.NewPreviewService(renderer, cache.NewPreviewMemoryCache(cfg.CacheTTL)))

	bannersHandler := handlers.NewBannersHandler(logger, previewUsecase)

	router.Get("/banners/preview/", bannersHandler.Preview)
	router.Post("/banners/", bannersHandler.Create)

	// Create and start HTTP server
	srv := server.New(cfg, router, logger)
	if err := srv.Start(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
