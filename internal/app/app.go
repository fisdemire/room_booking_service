package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fisdemire/room-booking/internal/config"
	"github.com/fisdemire/room-booking/internal/hash"
	jwtservice "github.com/fisdemire/room-booking/internal/jwt"
	"github.com/fisdemire/room-booking/internal/repository"
	"github.com/fisdemire/room-booking/internal/repository/postgres"
	"github.com/fisdemire/room-booking/internal/server"
	"github.com/fisdemire/room-booking/internal/usecase"
)

type App struct {
	repositories    *postgres.Repositories
	txManager       *postgres.TxManager
	server          *server.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

func New(
	ctx context.Context,
	cfg config.Config,
	log *slog.Logger,
) (*App, error) {
	repositories, txManager, err :=
		repository.NewRepositoryWithTxManager(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	tokenGenerator, err := jwtservice.NewService(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil {
		repositories.Close()
		return nil, fmt.Errorf("initialize JWT service: %w", err)
	}

	passwordHasher, err := hash.NewBcryptHasher(0)
	if err != nil {
		repositories.Close()
		return nil, fmt.Errorf("initialize password hasher: %w", err)
	}

	userUseCase := usecase.NewUserUseCase(
		repositories.User(),
		tokenGenerator,
		passwordHasher,
	)

	roomUseCase := usecase.NewRoomUseCase(
		repositories.Room(),
	)

	scheduleRepo := repositories.Schedule()
	slotRepo := repositories.Slot()

	slotGenerator := usecase.NewSlotGenerator(
		slotRepo,
		scheduleRepo,
	)

	slotUseCase := usecase.NewSlotUseCase(
		slotRepo,
		repositories.Room(),
	)

	scheduleUseCase := usecase.NewScheduleUseCase(
		scheduleRepo,
		slotGenerator,
		txManager,
	)

	srv := server.New(
		cfg.HTTP.GetAddr(),
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.HTTP.IdleTimeout,
		userUseCase,
		roomUseCase,
		scheduleUseCase,
		slotUseCase,
		tokenGenerator,
	)

	return &App{
		server:          srv,
		repositories:    repositories,
		txManager:       txManager,
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

	var runErr error

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")

	case err := <-serverErr:
		runErr = err
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		a.shutdownTimeout,
	)
	defer cancel()

	shutdownErr := a.Shutdown(shutdownCtx)

	return errors.Join(runErr, shutdownErr)
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down HTTP server")

	err := a.server.Shutdown(ctx)
	if err != nil {
		a.logger.Error(
			"graceful shutdown failed",
			"error",
			err,
		)

		err = errors.Join(err, a.server.Close())
	}

	a.repositories.Close()

	return err
}
