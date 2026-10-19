package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
)

type SlotWorker struct {
	generator domain.SlotGenerator
	interval  time.Duration
	daysAhead int
	logger    *slog.Logger
}

func NewSlotWorker(
	generator domain.SlotGenerator,
	interval time.Duration,
	daysAhead int,
	logger *slog.Logger,
) *SlotWorker {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	if daysAhead <= 0 {
		daysAhead = 7
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &SlotWorker{
		generator: generator,
		interval:  interval,
		daysAhead: daysAhead,
		logger:    logger,
	}
}

func (w *SlotWorker) Start(ctx context.Context) {
	w.logger.Info(
		"slot generation worker started",
		"interval", w.interval,
		"days_ahead", w.daysAhead,
	)

	if ctx.Err() != nil {
		return
	}

	w.generateSlots(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("slot generation worker stopped")
			return

		case <-ticker.C:
			w.generateSlots(ctx)
		}
	}
}

func (w *SlotWorker) generateSlots(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	w.logger.Info("generating slots", "days_ahead", w.daysAhead)

	if err := w.generator.GenerateSlots(ctx, w.daysAhead); err != nil {
		w.logger.Error("failed to generate slots", "error", err)
		return
	}

	w.logger.Info("slots generated successfully")
}
