package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hurtki/github-banners/storage/internal/config"
	"github.com/hurtki/github-banners/storage/internal/domain"
	bannersstorage "github.com/hurtki/github-banners/storage/internal/infrastructure/banners_storage"
	"github.com/hurtki/github-banners/storage/internal/logger"
)

func main() {
	config := config.Load()
	logger := logger.NewLogger(config.LogLevel, config.LogFormat)
	logger.Info("started storage service")

	bannersStorage := bannersstorage.NewFileStorage(config.BannersStoragePath, logger, os.WriteFile)
	err := bannersStorage.Save("alex-test", domain.SvgBannerExtension, []byte("test svg"))
	logger.Info("saved banner to storage", "err", err)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down, interrupt signal received")

	_, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
}
