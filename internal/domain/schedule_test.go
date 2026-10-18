package domain_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
)

func TestSchedule_NormalizeDays(t *testing.T) {
	schedule := domain.Schedule{
		DaysOfWeek: []int{5, 1, 3, 1, 7, 3},
	}

	schedule.NormalizeDays()

	want := []int{1, 3, 5, 7}

	if !slices.Equal(schedule.DaysOfWeek, want) {
		t.Errorf(
			"NormalizeDays() = %v, want %v",
			schedule.DaysOfWeek,
			want,
		)
	}
}

func TestSchedule_Validate(t *testing.T) {
	validSchedule := domain.Schedule{
		ID:         "schedule-1",
		RoomID:     "room-1",
		DaysOfWeek: []int{1, 3, 5},
		StartTime:  time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:    time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC),
	}

	tests := []struct {
		name    string
		mutate  func(*domain.Schedule)
		wantErr error
	}{
		{
			name: "valid schedule",
		},
		{
			name: "no days specified",
			mutate: func(s *domain.Schedule) {
				s.DaysOfWeek = nil
			},
			wantErr: domain.ErrNoDaysSpecified,
		},
		{
			name: "day zero",
			mutate: func(s *domain.Schedule) {
				s.DaysOfWeek = []int{0}
			},
			wantErr: domain.ErrInvalidDayOfWeek,
		},
		{
			name: "day eight",
			mutate: func(s *domain.Schedule) {
				s.DaysOfWeek = []int{8}
			},
			wantErr: domain.ErrInvalidDayOfWeek,
		},
		{
			name: "start equals end",
			mutate: func(s *domain.Schedule) {
				s.EndTime = s.StartTime
			},
			wantErr: domain.ErrInvalidTimeRange,
		},
		{
			name: "duration 29 minutes",
			mutate: func(s *domain.Schedule) {
				s.EndTime = s.StartTime.Add(29 * time.Minute)
			},
			wantErr: domain.ErrDurationTooShort,
		},
		{
			name: "duration exactly 30 minutes",
		},
		{
			name: "duration 45 minutes",
			mutate: func(s *domain.Schedule) {
				s.EndTime = s.StartTime.Add(45 * time.Minute)
			},
			wantErr: domain.ErrInvalidTimeRange,
		},
		{
			name: "different dates are ignored",
			mutate: func(s *domain.Schedule) {
				s.StartTime = time.Date(
					2026, 1, 1, 10, 0, 0, 0, time.UTC,
				)
				s.EndTime = time.Date(
					2026, 1, 2, 10, 30, 0, 0, time.UTC,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := validSchedule

			schedule.DaysOfWeek = slices.Clone(validSchedule.DaysOfWeek)

			if tt.mutate != nil {
				tt.mutate(&schedule)
			}

			err := schedule.Validate()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf(
					"Validate() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestSchedule_Weekdays(t *testing.T) {
	schedule := domain.Schedule{
		DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
	}

	got := schedule.Weekdays()

	want := []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
		time.Sunday,
	}

	if !slices.Equal(got, want) {
		t.Errorf("Weekdays() = %v, want %v", got, want)
	}
}
