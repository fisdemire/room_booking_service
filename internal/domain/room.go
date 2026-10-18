package domain

import (
	"context"
	"strings"
	"time"
)

type Room struct {
	ID          string
	Name        string
	Description string
	Capacity    int
	CreatedAt   time.Time
}

func (r *Room) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrInvalidRoomName
	}

	if r.Capacity < 1 {
		return ErrInvalidCapacity
	}

	return nil
}

type RoomRepository interface {
	Create(ctx context.Context, room *Room) error
	GetAll(ctx context.Context) ([]*Room, error)
}

type RoomUseCase interface {
	CreateRoom(
		ctx context.Context,
		name, description string,
		capacity int,
	) (*Room, error)

	GetAllRooms(ctx context.Context) ([]*Room, error)
}
