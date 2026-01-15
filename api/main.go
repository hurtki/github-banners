package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hurtki/github-banners/api/internal/cache"
	"github.com/hurtki/github-banners/api/internal/config"
	"github.com/hurtki/github-banners/api/internal/github"
	log "github.com/hurtki/github-banners/api/internal/logger"
)

func main() {
	cfg := config.Load()

	logger := log.NewLogger(cfg.LogLevel, cfg.LogFormat)
	logger.Info("Starting Github Stats API")

	memoryCache := cache.NewCache(cfg.CacheTTL)

	githubConfig := &github.ServiceConfig{
		CacheTTL: cfg.CacheTTL,
		RequestTimeout: cfg.RequestTimeout,
	}

	githubService := github.NewService(cfg.GithubToken, memoryCache, githubConfig)

	handler := github.NewHandler(cfg, logger, githubService)

	server := &http.Server {
		Addr: fmt.Sprintf(":%s", cfg.Port),
		Handler: handler, 
		ReadTimeout: 10*time.Second,
		WriteTimeout: 30*time.Second,
		IdleTimeout: 60*time.Second,
	}

	//graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down server")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Could not gracefully shutdown server", "error", err)
		}

		close(done)
	}()

	//start server
	logger.Info("Server is ready to handle requests", "port", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed{
		logger.Error("Could not listen on port", "port", cfg.Port, "error", err)
		os.Exit(1)
	}

	<-done
	logger.Info("Server stopped")
}
