package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/google/uuid"
)

const ScheduleDaysAhead = 7

const slotDuration = 30 * time.Minute

type slotGenerator struct {
	slotRepo     domain.SlotRepository
	scheduleRepo domain.ScheduleRepository
}

func NewSlotGenerator(
	slotRepo domain.SlotRepository,
	scheduleRepo domain.ScheduleRepository,
) domain.SlotGenerator {
	return &slotGenerator{
		slotRepo:     slotRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (g *slotGenerator) GenerateSlots(
	ctx context.Context,
	daysAhead int,
) error {
	if daysAhead < 1 {
		return fmt.Errorf("daysAhead must be positive")
	}

	schedules, err := g.scheduleRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get schedules: %w", err)
	}

	now := time.Now().UTC()
	startOfToday := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, time.UTC,
	)

	slots := make([]*domain.Slot, 0)

	for _, schedule := range schedules {
		if schedule == nil {
			continue
		}

		slots = append(
			slots,
			generateSlotsForSchedule(
				schedule,
				startOfToday,
				now,
				daysAhead,
			)...,
		)
	}

	if len(slots) == 0 {
		return nil
	}

	if err := g.slotRepo.CreateBulk(ctx, slots); err != nil {
		return fmt.Errorf("save generated slots: %w", err)
	}

	return nil
}

func (g *slotGenerator) GenerateSlotsForRoom(
	ctx context.Context,
	roomID string,
	daysAhead int,
) error {
	if daysAhead < 1 {
		return fmt.Errorf("daysAhead must be positive")
	}

	schedule, err := g.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("get schedule for room %s: %w", roomID, err)
	}
	if schedule == nil {
		return domain.ErrScheduleNotFound
	}

	now := time.Now().UTC()
	startOfToday := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, time.UTC,
	)

	slots := generateSlotsForSchedule(
		schedule,
		startOfToday,
		now,
		daysAhead,
	)

	if len(slots) == 0 {
		return nil
	}

	if err := g.slotRepo.CreateBulk(ctx, slots); err != nil {
		return fmt.Errorf("save generated slots: %w", err)
	}

	return nil
}

func generateSlotsForSchedule(
	schedule *domain.Schedule,
	startOfToday time.Time,
	now time.Time,
	daysAhead int,
) []*domain.Slot {
	weekdays := make(map[time.Weekday]struct{}, len(schedule.DaysOfWeek))
	for _, weekday := range schedule.Weekdays() {
		weekdays[weekday] = struct{}{}
	}

	slots := make([]*domain.Slot, 0)
	startTime := schedule.StartTime.UTC()
	endTime := schedule.EndTime.UTC()

	for i := 0; i < daysAhead; i++ {
		date := startOfToday.AddDate(0, 0, i)

		if _, ok := weekdays[date.Weekday()]; !ok {
			continue
		}

		dayStart := time.Date(
			date.Year(), date.Month(), date.Day(),
			startTime.Hour(), startTime.Minute(),
			0, 0, time.UTC,
		)
		dayEnd := time.Date(
			date.Year(), date.Month(), date.Day(),
			endTime.Hour(), endTime.Minute(),
			0, 0, time.UTC,
		)

		for start := dayStart; start.Before(dayEnd); start = start.Add(slotDuration) {
			end := start.Add(slotDuration)
			if end.After(dayEnd) || start.Before(now) {
				continue
			}

			slots = append(slots, &domain.Slot{
				ID:          uuid.NewString(),
				RoomID:      schedule.RoomID,
				StartTime:   start,
				EndTime:     end,
				IsAvailable: true,
			})
		}
	}

	return slots
}
