package main

import (
	"time"

	"github.com/hurtki/github-banners/api/internal/logger"
)

func main() {
	logger := logger.NewLogger()
	logger.Info("started api service")

	time.Sleep(time.Hour)

}
