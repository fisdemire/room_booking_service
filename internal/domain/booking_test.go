package domain_test

import (
	"testing"

	"github.com/fisdemire/room-booking/internal/domain"
)

func TestBooking_Cancel(t *testing.T) {
	booking := domain.Booking{
		ID:     "booking-1",
		SlotID: "slot-1",
		UserID: "user-1",
		Status: domain.BookingStatusActive,
	}

	if !booking.Cancel() {
		t.Fatal("Cancel() = false, want true for active booking")
	}

	if booking.Status != domain.BookingStatusCancelled {
		t.Errorf(
			"Status = %q, want %q",
			booking.Status,
			domain.BookingStatusCancelled,
		)
	}

	if string(booking.Status) != "cancelled" {
		t.Errorf(
			"status database value = %q, want cancelled",
			string(booking.Status),
		)
	}

	if booking.Cancel() {
		t.Error("second Cancel() = true, want false")
	}

	if booking.Status != domain.BookingStatusCancelled {
		t.Errorf(
			"Status after second cancellation = %q, want %q",
			booking.Status,
			domain.BookingStatusCancelled,
		)
	}
}

func TestBooking_IsOwnedBy(t *testing.T) {
	booking := domain.Booking{
		UserID: "user-1",
	}

	tests := []struct {
		name   string
		userID string
		want   bool
	}{
		{
			name:   "owner",
			userID: "user-1",
			want:   true,
		},
		{
			name:   "different user",
			userID: "user-2",
			want:   false,
		},
		{
			name:   "empty user ID",
			userID: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := booking.IsOwnedBy(tt.userID); got != tt.want {
				t.Errorf(
					"IsOwnedBy(%q) = %v, want %v",
					tt.userID,
					got,
					tt.want,
				)
			}
		})
	}
}
