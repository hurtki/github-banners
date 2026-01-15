package main

import (
	"time"

	"github.com/hurtki/github-banners/renderer/internal/logger"
)

func main() {
	logger := logger.NewLogger()
	logger.Info("started renderer service")
	time.Sleep(time.Hour)

}
