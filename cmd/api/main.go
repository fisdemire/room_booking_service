package main

import (
	"log/slog"
	"os"

	"github.com/fisdemire/room-booking/internal/config"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
	slog.SetDefault(logger)

	cfg, err := config.New()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
}
