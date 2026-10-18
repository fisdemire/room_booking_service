package domain

import (
	"context"
	"slices"
	"time"
)

type Schedule struct {
	ID         string
	RoomID     string
	DaysOfWeek []int
	StartTime  time.Time
	EndTime    time.Time
}

func (s *Schedule) NormalizeDays() {
	slices.Sort(s.DaysOfWeek)
	s.DaysOfWeek = slices.Compact(s.DaysOfWeek)
}

func (s *Schedule) Weekdays() []time.Weekday {
	weekdays := make([]time.Weekday, 0, len(s.DaysOfWeek))

	for _, day := range s.DaysOfWeek {
		if day < 1 || day > 7 {
			continue
		}

		weekdays = append(weekdays, time.Weekday(day%7))
	}

	return weekdays
}

func (s *Schedule) Validate() error {
	if len(s.DaysOfWeek) == 0 {
		return ErrNoDaysSpecified
	}

	for _, day := range s.DaysOfWeek {
		if day < 1 || day > 7 {
			return ErrInvalidDayOfWeek
		}
	}

	start := minutesSinceMidnight(s.StartTime)
	end := minutesSinceMidnight(s.EndTime)

	if start >= end {
		return ErrInvalidTimeRange
	}

	duration := end - start

	if duration < 30 {
		return ErrDurationTooShort
	}

	if duration%30 != 0 {
		return ErrInvalidTimeRange
	}

	return nil
}

func minutesSinceMidnight(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *Schedule) (*Schedule, error)
	GetByRoomID(ctx context.Context, roomID string) (*Schedule, error)
	GetAll(ctx context.Context) ([]*Schedule, error)
}

type ScheduleUseCase interface {
	CreateSchedule(
		ctx context.Context,
		roomID string,
		daysOfWeek []int,
		startTime, endTime string,
	) (*Schedule, error)
}
