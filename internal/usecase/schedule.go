package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/google/uuid"
)

type TransactionRunner interface {
	WithinTransaction(
		ctx context.Context,
		fn func(context.Context) error,
	) error
}

type scheduleUseCase struct {
	scheduleRepo  domain.ScheduleRepository
	slotGenerator domain.SlotGenerator
	txManager     TransactionRunner
}

func NewScheduleUseCase(
	scheduleRepo domain.ScheduleRepository,
	slotGenerator domain.SlotGenerator,
	txManagers ...TransactionRunner,
) domain.ScheduleUseCase {
	var txManager TransactionRunner
	if len(txManagers) > 0 {
		txManager = txManagers[0]
	}

	return &scheduleUseCase{
		scheduleRepo:  scheduleRepo,
		slotGenerator: slotGenerator,
		txManager:     txManager,
	}
}

func (uc *scheduleUseCase) CreateSchedule(
	ctx context.Context,
	roomID string,
	daysOfWeek []int,
	startTime string,
	endTime string,
) (*domain.Schedule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: invalid start time",
			domain.ErrInvalidTimeRange,
		)
	}

	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: invalid end time",
			domain.ErrInvalidTimeRange,
		)
	}

	schedule := &domain.Schedule{
		ID:         uuid.NewString(),
		RoomID:     roomID,
		DaysOfWeek: append([]int(nil), daysOfWeek...),
		StartTime:  start.UTC(),
		EndTime:    end.UTC(),
	}
	schedule.NormalizeDays()

	if err := schedule.Validate(); err != nil {
		return nil, fmt.Errorf("validate schedule: %w", err)
	}

	var created *domain.Schedule

	create := func(txCtx context.Context) error {
		saved, err := uc.scheduleRepo.Create(txCtx, schedule)
		if err != nil {
			return fmt.Errorf("save schedule: %w", err)
		}
		if saved == nil {
			return fmt.Errorf("save schedule: repository returned nil")
		}

		if err := uc.slotGenerator.GenerateSlotsForRoom(
			txCtx,
			saved.RoomID,
			ScheduleDaysAhead,
		); err != nil {
			return fmt.Errorf("generate slots: %w", err)
		}

		created = saved
		return nil
	}

	if uc.txManager != nil {
		err = uc.txManager.WithinTransaction(ctx, create)
	} else {
		err = create(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("create schedule: %w", err)
	}

	return created, nil
}
