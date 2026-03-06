package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hurtki/github-banners/storage/internal/config"
	"github.com/hurtki/github-banners/storage/internal/domain/banner"
	"github.com/hurtki/github-banners/storage/internal/handlers"
	bannersstorage "github.com/hurtki/github-banners/storage/internal/infrastructure/banners_storage"
	"github.com/hurtki/github-banners/storage/internal/infrastructure/server"
	"github.com/hurtki/github-banners/storage/internal/logger"
)

func main() {
	config := config.Load()
	logger := logger.NewLogger(config.LogLevel, config.LogFormat)
	logger.Info("started storage service")

	bannersStorage := bannersstorage.NewFileStorage(config.BannersStoragePath, logger, os.WriteFile)
	usecase := banner.NewBannerUsecase(bannersStorage)
	handler := handlers.NewBannerSaveHandler(logger, usecase)

	router := chi.NewRouter()
	router.Post("/banners", handler.Save)
	srv := server.New(config, router, logger)
	srv.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down, interrupt signal received")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	srv.Close(ctx)

	defer cancel()
}
