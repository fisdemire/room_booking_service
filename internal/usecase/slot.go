package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
)

type slotUseCase struct {
	slotRepo domain.SlotRepository
	roomRepo domain.RoomRepository
}

func NewSlotUseCase(
	slotRepo domain.SlotRepository,
	roomRepo domain.RoomRepository,
) domain.SlotUseCase {
	return &slotUseCase{
		slotRepo: slotRepo,
		roomRepo: roomRepo,
	}
}

func (uc *slotUseCase) GetAvailableSlots(
	ctx context.Context,
	roomID string,
	date time.Time,
) ([]*domain.Slot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, err := uc.roomRepo.GetByID(ctx, roomID); err != nil {
		return nil, fmt.Errorf("find room: %w", err)
	}

	slots, err := uc.slotRepo.GetAvailableByRoomAndTime(
		ctx,
		roomID,
		date.UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("get available slots: %w", err)
	}

	now := time.Now().UTC()
	available := make([]*domain.Slot, 0, len(slots))

	for _, slot := range slots {
		if slot == nil {
			continue
		}

		if !slot.StartTime.Before(now) && slot.IsAvailable {
			available = append(available, slot)
		}
	}

	return available, nil
}

func (uc *slotUseCase) CancelSlot(
	ctx context.Context,
	roomID string,
	slotID string,
) error {
	if roomID == "" || slotID == "" {
		return domain.ErrRoomIDAndSlotIDRequired
	}

	slot, err := uc.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("get slot: %w", err)
	}

	if slot.RoomID != roomID {
		return domain.ErrSlotNotInRoom
	}

	slot.Cancel()

	if err := uc.slotRepo.Update(ctx, slot); err != nil {
		return fmt.Errorf("restore slot availability: %w", err)
	}

	return nil
}
