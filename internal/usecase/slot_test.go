package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/usecase"
	"github.com/fisdemire/room-booking/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSlotUseCase_GetAvailableSlotsFiltersPast(t *testing.T) {
	ctx := context.Background()
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	roomRepo := mocks.NewMockRoomRepository(t)
	slotRepo := mocks.NewMockSlotRepository(t)

	roomRepo.EXPECT().
		GetByID(mock.Anything, roomID).
		Return(&domain.Room{ID: roomID}, nil).
		Once()

	now := time.Now().UTC()

	futureSlot := &domain.Slot{
		ID:          "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		RoomID:      roomID,
		StartTime:   now.Add(2 * time.Hour),
		EndTime:     now.Add(150 * time.Minute),
		IsAvailable: true,
	}

	pastSlot := &domain.Slot{
		ID:          "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		RoomID:      roomID,
		StartTime:   now.Add(-2 * time.Hour),
		EndTime:     now.Add(-90 * time.Minute),
		IsAvailable: true,
	}

	date := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, time.UTC,
	)

	slotRepo.EXPECT().
		GetAvailableByRoomAndTime(mock.Anything, roomID, date).
		Return([]*domain.Slot{pastSlot, futureSlot}, nil).
		Once()

	uc := usecase.NewSlotUseCase(slotRepo, roomRepo)

	slots, err := uc.GetAvailableSlots(ctx, roomID, date)

	require.NoError(t, err)
	require.Len(t, slots, 1)
	require.Equal(t, futureSlot.ID, slots[0].ID)
}

func TestSlotUseCase_GetAvailableSlotsEmpty(t *testing.T) {
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	roomRepo := mocks.NewMockRoomRepository(t)
	slotRepo := mocks.NewMockSlotRepository(t)

	roomRepo.EXPECT().
		GetByID(mock.Anything, roomID).
		Return(&domain.Room{ID: roomID}, nil).
		Once()

	date := time.Now().UTC()
	date = time.Date(
		date.Year(), date.Month(), date.Day(),
		0, 0, 0, 0, time.UTC,
	)

	slotRepo.EXPECT().
		GetAvailableByRoomAndTime(mock.Anything, roomID, date).
		Return(nil, nil).
		Once()

	uc := usecase.NewSlotUseCase(slotRepo, roomRepo)

	slots, err := uc.GetAvailableSlots(context.Background(), roomID, date)

	require.NoError(t, err)
	require.NotNil(t, slots)
	require.Empty(t, slots)
}

func TestSlotUseCase_RoomNotFound(t *testing.T) {
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	roomRepo := mocks.NewMockRoomRepository(t)
	slotRepo := mocks.NewMockSlotRepository(t)

	roomRepo.EXPECT().
		GetByID(mock.Anything, roomID).
		Return(nil, domain.ErrRoomNotFound).
		Once()

	uc := usecase.NewSlotUseCase(slotRepo, roomRepo)

	slots, err := uc.GetAvailableSlots(
		context.Background(),
		roomID,
		time.Now().UTC(),
	)

	require.Nil(t, slots)
	require.ErrorIs(t, err, domain.ErrRoomNotFound)
}
