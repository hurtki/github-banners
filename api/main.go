package main

import (
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/domain"
	"github.com/hurtki/github-banners/api/internal/domain/services"
	"github.com/hurtki/github-banners/api/internal/handlers"
	infraDB "github.com/hurtki/github-banners/api/internal/infrastructure/db"
	infraGithub "github.com/hurtki/github-banners/api/internal/infrastructure/github"
	"github.com/hurtki/github-banners/api/internal/infrastructure/renderer"
	renderer_http "github.com/hurtki/github-banners/api/internal/infrastructure/renderer/http"
	"github.com/hurtki/github-banners/api/internal/infrastructure/server"
	log "github.com/hurtki/github-banners/api/internal/logger"
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

	// renderer infra intialization
	rendererAithRT := renderer_http.NewRendererAuthHTTPRoundTripper("api", renderer_http.NewHMACSigner([]byte(cfg.ServicesSecret)), time.Now)
	rendererHTTPClient := renderer_http.NewRendererHTTPClient(rendererAithRT)
	/*rendererClient */ _ = renderer.NewRenderer(rendererHTTPClient, logger, "https://renderer/preview/")

	// Create service configuration
	serviceConfig := &domain.ServiceConfig{
		CacheTTL:       cfg.CacheTTL,
		RequestTimeout: cfg.RequestTimeout,
	}
	// Create GitHub fetcher (infrastructure layer)
	githubFetcher := infraGithub.NewFetcher(cfg.GithubToken, serviceConfig)

	// Create stats service (domain service with cache)
	statsService := services.NewStatsService(githubFetcher, memoryCache)
	db, err := infraDB.NewDB(psgrConf, logger)
	if err != nil {
		logger.Error("can't intialize database, exiting", "err", err.Error())
		os.Exit(1)
	}

	// temp usage to compile
	db.Stats()

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
