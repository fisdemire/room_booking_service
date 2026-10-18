package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
)

func TestSlot_IsPast(t *testing.T) {
	now := time.Date(
		2026, 1, 1, 12, 0, 0, 0, time.UTC,
	)

	tests := []struct {
		name      string
		startTime time.Time
		want      bool
	}{
		{
			name:      "slot in past",
			startTime: now.Add(-time.Second),
			want:      true,
		},
		{
			name:      "slot starts exactly now",
			startTime: now,
			want:      false,
		},
		{
			name:      "slot in future",
			startTime: now.Add(time.Second),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot := domain.Slot{
				StartTime: tt.startTime,
			}

			if got := slot.IsPast(now); got != tt.want {
				t.Errorf(
					"IsPast() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestSlot_Book(t *testing.T) {
	slot := domain.Slot{
		ID:          "slot-1",
		RoomID:      "room-1",
		IsAvailable: true,
	}

	if err := slot.Book(); err != nil {
		t.Fatalf("Book() error = %v, want nil", err)
	}

	if slot.IsAvailable {
		t.Error("slot should become unavailable after booking")
	}

	err := slot.Book()

	if !errors.Is(err, domain.ErrSlotOccupied) {
		t.Errorf(
			"second Book() error = %v, want %v",
			err,
			domain.ErrSlotOccupied,
		)
	}
}

func TestSlot_CancelAndBookAgain(t *testing.T) {
	slot := domain.Slot{
		IsAvailable: true,
	}

	if err := slot.Book(); err != nil {
		t.Fatalf("first Book() error = %v", err)
	}

	slot.Cancel()

	if !slot.IsAvailable {
		t.Fatal("slot should become available after cancellation")
	}

	if err := slot.Book(); err != nil {
		t.Fatalf("second booking after cancellation failed: %v", err)
	}

	if slot.IsAvailable {
		t.Error("slot should become unavailable after rebooking")
	}
}

func TestSlot_Disable(t *testing.T) {
	slot := domain.Slot{
		IsAvailable: true,
	}

	slot.Disable()

	if slot.IsAvailable {
		t.Fatal("slot should be unavailable after Disable()")
	}

	if err := slot.Book(); !errors.Is(err, domain.ErrSlotOccupied) {
		t.Errorf("Book() error = %v, want %v", err, domain.ErrSlotOccupied)
	}
}
