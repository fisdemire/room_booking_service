package domain

import (
	"context"
	"time"
)

type Slot struct {
	ID          string
	RoomID      string
	StartTime   time.Time
	EndTime     time.Time
	IsAvailable bool
}

func (s *Slot) IsPast(now time.Time) bool {
	return s.StartTime.Before(now)
}

func (s *Slot) Book() error {
	if !s.IsAvailable {
		return ErrSlotOccupied
	}

	s.IsAvailable = false

	return nil
}

func (s *Slot) Cancel() {
	s.IsAvailable = true
}

func (s *Slot) Disable() {
	s.IsAvailable = false
}

type SlotRepository interface {
	CreateBulk(ctx context.Context, slots []*Slot) error
	GetAvailableByRoomAndTime(
		ctx context.Context,
		roomID string,
		date time.Time,
	) ([]*Slot, error)
	GetByID(ctx context.Context, slotID string) (*Slot, error)
	Update(ctx context.Context, slot *Slot) error
}

type SlotUseCase interface {
	GetAvailableSlots(
		ctx context.Context,
		roomID string,
		date time.Time,
	) ([]*Slot, error)

	CancelSlot(ctx context.Context, roomID, slotID string) error
}

type SlotGenerator interface {
	GenerateSlots(ctx context.Context, daysAhead int) error
	GenerateSlotsForRoom(
		ctx context.Context,
		roomID string,
		daysAhead int,
	) error
}

type SlotService interface {
	SlotUseCase
	SlotGenerator
}
