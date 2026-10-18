package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/usecase"
	"github.com/fisdemire/room-booking/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScheduleUseCase_CreateScheduleValidation(t *testing.T) {
	tests := []struct {
		name      string
		days      []int
		startTime string
		endTime   string
		wantErr   error
	}{
		{
			name:      "no days",
			days:      nil,
			startTime: "09:00",
			endTime:   "18:00",
			wantErr:   domain.ErrNoDaysSpecified,
		},
		{
			name:      "invalid day",
			days:      []int{8},
			startTime: "09:00",
			endTime:   "18:00",
			wantErr:   domain.ErrInvalidDayOfWeek,
		},
		{
			name:      "invalid start time",
			days:      []int{1},
			startTime: "not-time",
			endTime:   "18:00",
			wantErr:   domain.ErrInvalidTimeRange,
		},
		{
			name:      "duration below 30 minutes",
			days:      []int{1},
			startTime: "09:00",
			endTime:   "09:15",
			wantErr:   domain.ErrDurationTooShort,
		},
		{
			name:      "duration not divisible by 30",
			days:      []int{1},
			startTime: "09:00",
			endTime:   "09:45",
			wantErr:   domain.ErrInvalidTimeRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockScheduleRepository(t)
			generator := mocks.NewMockSlotGenerator(t)

			uc := usecase.NewScheduleUseCase(repo, generator)

			schedule, err := uc.CreateSchedule(
				context.Background(),
				"550e8400-e29b-41d4-a716-446655440000",
				tt.days,
				tt.startTime,
				tt.endTime,
			)

			require.Nil(t, schedule)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestScheduleUseCase_CreateScheduleSuccess(t *testing.T) {
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	repo := mocks.NewMockScheduleRepository(t)
	generator := mocks.NewMockSlotGenerator(t)

	repo.EXPECT().
		Create(
			mock.Anything,
			mock.MatchedBy(func(schedule *domain.Schedule) bool {
				if schedule == nil {
					return false
				}

				_, err := uuid.Parse(schedule.ID)

				return err == nil &&
					schedule.RoomID == roomID &&
					len(schedule.DaysOfWeek) == 2 &&
					schedule.DaysOfWeek[0] == 1 &&
					schedule.DaysOfWeek[1] == 5 &&
					schedule.StartTime.Format("15:04") == "09:00" &&
					schedule.EndTime.Format("15:04") == "18:00"
			}),
		).
		RunAndReturn(func(
			_ context.Context,
			schedule *domain.Schedule,
		) (*domain.Schedule, error) {
			return schedule, nil
		}).
		Once()

	generator.EXPECT().
		GenerateSlotsForRoom(
			mock.Anything,
			roomID,
			usecase.ScheduleDaysAhead,
		).
		Return(nil).
		Once()

	uc := usecase.NewScheduleUseCase(repo, generator)

	schedule, err := uc.CreateSchedule(
		context.Background(),
		roomID,
		[]int{5, 1, 1},
		"09:00",
		"18:00",
	)

	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.Equal(t, []int{1, 5}, schedule.DaysOfWeek)
	require.Equal(t, "09:00", schedule.StartTime.Format("15:04"))
	require.Equal(t, "18:00", schedule.EndTime.Format("15:04"))
	require.Equal(t, time.UTC, schedule.StartTime.Location())
}

func TestScheduleUseCase_ScheduleAlreadyExists(t *testing.T) {
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	repo := mocks.NewMockScheduleRepository(t)
	generator := mocks.NewMockSlotGenerator(t)

	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil, domain.ErrScheduleAlreadyExists).
		Once()

	uc := usecase.NewScheduleUseCase(repo, generator)

	schedule, err := uc.CreateSchedule(
		context.Background(),
		roomID,
		[]int{1, 2, 3, 4, 5},
		"09:00",
		"18:00",
	)

	require.Nil(t, schedule)
	require.ErrorIs(t, err, domain.ErrScheduleAlreadyExists)
}
