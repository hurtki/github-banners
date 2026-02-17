package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	kafka_config "github.com/hurtki/github-banners/renderer/internal/config/kafka"
	"github.com/hurtki/github-banners/renderer/internal/handlers"
	"github.com/hurtki/github-banners/renderer/internal/infrastructure/kafka"
	kafka_cg_handlers "github.com/hurtki/github-banners/renderer/internal/infrastructure/kafka/cg_handlers"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logger.NewLogger("info", "json")
	logger.Info("started renderer service")
	bannerUpdateHandler := handlers.NewBannerUpdateHandler(logger)

	cgHandlerCfg := kafka_config.NewKafkaCGHandlerConfig()

	cgBannerUpdateHandler := kafka_cg_handlers.NewBannerUpdateCGHandler(logger, bannerUpdateHandler, cgHandlerCfg)
	kafkaConsumerCfg := kafka_config.NewKafkaConsumerConfig()

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
