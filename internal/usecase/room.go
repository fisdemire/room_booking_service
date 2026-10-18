package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/google/uuid"
)

type roomUseCase struct {
	roomRepo domain.RoomRepository
}

func NewRoomUseCase(
	roomRepo domain.RoomRepository,
) domain.RoomUseCase {
	return &roomUseCase{
		roomRepo: roomRepo,
	}
}

func (uc *roomUseCase) CreateRoom(
	ctx context.Context,
	name string,
	description string,
	capacity int,
) (*domain.Room, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	room := &domain.Room{
		ID:          uuid.New().String(),
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Capacity:    capacity,
		CreatedAt:   time.Now().UTC(),
	}

	if err := room.Validate(); err != nil {
		return nil, fmt.Errorf("validate room: %w", err)
	}

	if err := uc.roomRepo.Create(ctx, room); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	return room, nil
}

func (uc *roomUseCase) GetAllRooms(
	ctx context.Context,
) ([]*domain.Room, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	rooms, err := uc.roomRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all rooms: %w", err)
	}

	if rooms == nil {
		rooms = make([]*domain.Room, 0)
	}

	return rooms, nil
}
