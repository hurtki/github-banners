package main

import (
	"time"

	"github.com/hurtki/github-banners/storage/internal/logger"
)

func main() {
	logger := logger.NewLogger()
	logger.Info("started storage service")
	time.Sleep(time.Hour)
}
