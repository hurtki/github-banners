package logger

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	options := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, &options)

	return slog.New(handler)
}
