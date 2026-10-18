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

func TestRoomUseCase_CreateRoomInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		roomName string
		capacity int
		wantErr  error
	}{
		{
			name:     "empty name",
			roomName: "   ",
			capacity: 4,
			wantErr:  domain.ErrInvalidRoomName,
		},
		{
			name:     "zero capacity",
			roomName: "Loft",
			capacity: 0,
			wantErr:  domain.ErrInvalidCapacity,
		},
		{
			name:     "negative capacity",
			roomName: "Loft",
			capacity: -1,
			wantErr:  domain.ErrInvalidCapacity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRoomRepository(t)
			uc := usecase.NewRoomUseCase(repo)

			room, err := uc.CreateRoom(
				context.Background(),
				tt.roomName,
				"",
				tt.capacity,
			)

			require.Nil(t, room)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoomUseCase_CreateRoomSuccess(t *testing.T) {
	repo := mocks.NewMockRoomRepository(t)

	repo.EXPECT().
		Create(
			mock.Anything,
			mock.MatchedBy(func(room *domain.Room) bool {
				if room == nil {
					return false
				}

				_, err := uuid.Parse(room.ID)

				return err == nil &&
					room.Name == "Loft" &&
					room.Description == "Small room" &&
					room.Capacity == 4 &&
					!room.CreatedAt.IsZero() &&
					room.CreatedAt.Location() == time.UTC
			}),
		).
		Return(nil).
		Once()

	uc := usecase.NewRoomUseCase(repo)

	room, err := uc.CreateRoom(
		context.Background(),
		" Loft ",
		" Small room ",
		4,
	)

	require.NoError(t, err)
	require.NotNil(t, room)
	require.Equal(t, "Loft", room.Name)
	require.Equal(t, "Small room", room.Description)
	require.Equal(t, 4, room.Capacity)
	require.NotEmpty(t, room.ID)
}

func TestRoomUseCase_GetAllRoomsEmpty(t *testing.T) {
	repo := mocks.NewMockRoomRepository(t)

	repo.EXPECT().
		GetAll(mock.Anything).
		Return(nil, nil).
		Once()

	uc := usecase.NewRoomUseCase(repo)

	rooms, err := uc.GetAllRooms(context.Background())

	require.NoError(t, err)
	require.NotNil(t, rooms)
	require.Empty(t, rooms)
}
