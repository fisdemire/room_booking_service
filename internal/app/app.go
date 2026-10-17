package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fisdemire/room-booking/internal/config"
	"github.com/fisdemire/room-booking/internal/server"
)

type App struct {
	server          *server.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	srv := server.New(
		cfg.HTTP.GetAddr(),
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.HTTP.IdleTimeout,
	)

	return &App{
		server:          srv,
		logger:          log,
		shutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- a.server.Run()
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")

	case err := <-serverErr:
		if err != nil {
			return err
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		a.shutdownTimeout,
	)
	defer cancel()

	return a.Shutdown(shutdownCtx)
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down HTTP server")

	err := a.server.Shutdown(ctx)
	if err == nil {
		return nil
	}

	a.logger.Error("graceful shutdown failed", "error", err)

	if closeErr := a.server.Close(); closeErr != nil {
		return errors.Join(err, closeErr)
	}

	return err
}
