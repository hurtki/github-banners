package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hurtki/github-banners/renderer/internal/config"
	"github.com/hurtki/github-banners/renderer/internal/handlers"
	"github.com/hurtki/github-banners/renderer/internal/infrastructure/clients/storage"
	httpauth "github.com/hurtki/github-banners/renderer/internal/infrastructure/httpauth"
	"github.com/hurtki/github-banners/renderer/internal/infrastructure/kafka"
	kafka_cg_handlers "github.com/hurtki/github-banners/renderer/internal/infrastructure/kafka/cg_handlers"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	logger := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("started renderer service")

	signer := httpauth.NewHMACSigner([]byte(cfg.ServiceSecret))

	authTripper := httpauth.NewAuthHTTPRoundTripper("renderer-ms", signer, time.Now)
	httpClient := &http.Client{
		Transport : authTripper,
		Timeout : time.Second * 15,
	}

	storageClient := storage.NewClient(cfg.StorageBaseURL, httpClient, logger)
	bannerUpdateHandler := handlers.NewBannerUpdateHandler(logger, storageClient)

	cgHandlerCfg := config.NewKafkaCGHandlerConfig()

	cgBannerUpdateHandler := kafka_cg_handlers.NewBannerUpdateCGHandler(logger, bannerUpdateHandler, cgHandlerCfg)
	kafkaConsumerCfg := config.NewKafkaConsumerConfig()

	cs, err := kafka.NewKafkaConsumerGroup(ctx, logger, kafkaConsumerCfg)
	if err != nil {
		logger.Error("can't initialize kafka consumer group", "err", err)
		os.Exit(1)
	}

	cs.RegisterCGHandler([]string{"banner-update"}, cgBannerUpdateHandler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down, interrupt signal received")
	// base context cancel in defer
}
