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
	cfg := config.Load()

	logger := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("started renderer service")

	signer := httpauth.NewHMACSigner([]byte(cfg.ServiceSecret))

	authTripper := httpauth.NewAuthHTTPRoundTripper("renderer-ms", signer, time.Now)
	httpClient := &http.Client{
		Transport: authTripper,
		Timeout:   time.Second * 15,
	}

	_ = storage.NewClient(cfg.StorageBaseURL, httpClient, logger)
	bannerUpdateHandler := handlers.NewBannerUpdateHandler(logger)

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
