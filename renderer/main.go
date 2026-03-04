package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hurtki/github-banners/renderer/internal/config"
	"github.com/hurtki/github-banners/renderer/internal/domain/render"
	"github.com/hurtki/github-banners/renderer/internal/domain/templates"
	"github.com/hurtki/github-banners/renderer/internal/handlers/events"
	http_handlers "github.com/hurtki/github-banners/renderer/internal/handlers/http"
	"github.com/hurtki/github-banners/renderer/internal/infrastructure/clients/storage"
	httpauth "github.com/hurtki/github-banners/renderer/internal/infrastructure/httpauth"
	"github.com/hurtki/github-banners/renderer/internal/infrastructure/kafka"
	kafka_cg_handlers "github.com/hurtki/github-banners/renderer/internal/infrastructure/kafka/cg_handlers"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

func main() {
	cfg := config.Load()

	logger := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("started renderer service")

	signer := httpauth.NewHMACSigner([]byte(cfg.ServiceSecret))

	authTripper := httpauth.NewAuthHTTPRoundTripper("renderer-ms", signer, time.Now)
	httpClient := &http.Client{
		Transport: authTripper,
		Timeout:   time.Second * 15,
	}

	storageClient := storage.NewClient(cfg.StorageBaseURL, httpClient, logger)

	renderer, err := templates.NewRenderer()
	if err != nil {
		logger.Error("can't initialize renderer templates", "err", err)
		os.Exit(1)
	}

	renderUsecase := render.NewUsecase(renderer, storageClient)

	bannerUpdateHandler := events.NewBannerUpdateHandler(logger, renderUsecase)

	previewHandler := http_handlers.NewPreviewHandler(logger, renderUsecase)

	router := chi.NewRouter()
	router.Post("/preview", previewHandler.Preview)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		logger.Info("starting HTTP server", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "err", err)
		}
	}()

	cgHandlerCfg := config.NewKafkaCGHandlerConfig()

	cgBannerUpdateHandler := kafka_cg_handlers.NewBannerUpdateCGHandler(logger, bannerUpdateHandler, cgHandlerCfg)
	kafkaConsumerCfg := config.NewKafkaConsumerConfig()

	cg, err := kafka.NewKafkaConsumerGroup(logger, kafkaConsumerCfg)
	if err != nil {
		logger.Error("can't initialize kafka consumer group", "err", err)
		os.Exit(1)
	}

	cg.RegisterCGHandler([]string{"banner-update"}, cgBannerUpdateHandler)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down, interrupt signal received")

	quitCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// close consumer group
	cg.Close(quitCtx)
}
