package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/fisdemire/room-booking/internal/app"
	"github.com/fisdemire/room-booking/internal/config"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	cfg, err := config.New()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	application, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("initialize application", "error", err)
		os.Exit(1)
	}

	if err := application.Run(context.Background()); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}
