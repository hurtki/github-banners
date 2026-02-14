package main

import (
	"context"
	"os"

	"github.com/IBM/sarama"
	"github.com/hurtki/github-banners/renderer/internal/config"
	"github.com/hurtki/github-banners/renderer/internal/handlers"
	"github.com/hurtki/github-banners/renderer/internal/infrastrcture/kafka"
	"github.com/hurtki/github-banners/renderer/internal/logger"
)

func main() {
	logger := logger.NewLogger("info", "json")
	logger.Info("started renderer service")
	bannerUpdateHandler := handlers.NewBannerUpdateHandler(logger)
	cgBannerUpdateHandler := kafka.NewConsumeGroupHandler(logger, bannerUpdateHandler)
	kafkaCfg := config.KafkaConsumerConfig{Addrs: []string{"kafka:9092"}, SaramaCfg: sarama.NewConfig()}
	cs, err := kafka.NewKafkaConsumer(logger, kafkaCfg, cgBannerUpdateHandler)

	if err != nil {
		logger.Error("can't init kafka consumer", "err", err)
		os.Exit(1)
	}
	cs.StartListening(context.Background())
}
