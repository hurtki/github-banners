package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	banners_worker "github.com/hurtki/github-banners/api/internal/app/banners"
	user_stats_worker "github.com/hurtki/github-banners/api/internal/app/user_stats"
	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	longterm "github.com/hurtki/github-banners/api/internal/domain/long-term"
	"github.com/hurtki/github-banners/api/internal/domain/preview"
	userstats "github.com/hurtki/github-banners/api/internal/domain/user_stats"
	"github.com/hurtki/github-banners/api/internal/handlers"
	infraDB "github.com/hurtki/github-banners/api/internal/infrastructure/db"
	infraGithub "github.com/hurtki/github-banners/api/internal/infrastructure/github"
	http_auth "github.com/hurtki/github-banners/api/internal/infrastructure/httpauth"
	"github.com/hurtki/github-banners/api/internal/infrastructure/kafka"
	"github.com/hurtki/github-banners/api/internal/infrastructure/renderer"
	renderer_http "github.com/hurtki/github-banners/api/internal/infrastructure/renderer/http"
	"github.com/hurtki/github-banners/api/internal/infrastructure/server"
	"github.com/hurtki/github-banners/api/internal/infrastructure/storage"
	"github.com/hurtki/github-banners/api/internal/logger"
	"github.com/hurtki/github-banners/api/internal/migrations"
	banners_repo "github.com/hurtki/github-banners/api/internal/repo/banners"
	github_data_repo "github.com/hurtki/github-banners/api/internal/repo/github_user_data"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting github banners API service")

	psgrConf, err := config.LoadPostgres()
	if err != nil {
		logger.Error("Can't load postgres database config", "err", err.Error())
		os.Exit(1)

	}

	// cache
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
	repo := github_data_repo.NewGithubDataPsgrRepo(db, logger)

	// Create stats service (domain service with cache)
	statsService := userstats.NewUserStatsService(repo, githubFetcher, statsCache)

	router := chi.NewRouter()

	// renderer infra intialization
	rendererAuthRT := http_auth.NewAuthHTTPRoundTripper("api", http_auth.NewHMACSigner([]byte(cfg.ServicesSecret)), time.Now)
	rendererHTTPClient := renderer_http.NewRendererHTTPClient(rendererAuthRT)
	rendererCl := renderer.NewRenderer(rendererHTTPClient, logger, cfg.RendererBaseURL)

	// storage infra initialization
	storageAuthRT := http_auth.NewAuthHTTPRoundTripper(
		"api",
		http_auth.NewHMACSigner([]byte(cfg.ServicesSecret)),
		time.Now,
	)

	storageHTTPClient := &http.Client{
		Transport: storageAuthRT,
	}

	storageCl := storage.NewClient(
		cfg.StorageBaseURL,
		storageHTTPClient,
		logger,
	)

	previewService := preview.NewPreviewService(rendererCl, cache.NewPreviewMemoryCache(cfg.CacheTTL))

	previewUsecase := preview.NewPreviewUsecase(statsService, previewService)

	kafkaProducer, err := kafka.NewBannerProducer([]string{"kafka:9092"}, "banner-update", config.NewProducerConfig(), logger)
	if err != nil {
		logger.Error("can't connect to kafka as a producer", "err", err)
		os.Exit(1)
	}

	bannersRepo := banners_repo.NewPostgresRepo(db, logger)

	ltBannersUsecase := longterm.NewLTBannersUsecase(
		bannersRepo,
		kafkaProducer,
		previewService,
		storageCl,
		statsService,
	)

	bannersHandler := handlers.NewBannersHandler(logger, previewUsecase, ltBannersUsecase)

	// http handlers
	router.Get("/banners/preview", bannersHandler.Preview)
	router.Post("/banners", bannersHandler.Create)

	// workers startup
	ltBannersUpdateWorker := banners_worker.NewBannersWorker(logger, ltBannersUsecase.UpdateAll, time.Hour, longterm.UpdateAllConfig{Concurrency: 20})
	statsWorker := user_stats_worker.NewStatsWorker(statsService.RefreshAll, time.Hour, logger, userstats.WorkerConfig{BatchSize: 5, Concurrency: 10})

	ltBannersUpdateWorker.Start()
	statsWorker.Start()

	// Create and start HTTP server
	srv := server.New(cfg, router, logger)
	srv.Start()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	quitCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	ltBannersUpdateWorker.Close(quitCtx)
	statsWorker.Close(quitCtx)
	srv.Close(quitCtx)

	cancel()
}
