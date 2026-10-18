package domain

import (
	"context"
	"time"
)

type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID             string
	SlotID         string
	UserID         string
	Status         BookingStatus
	ConferenceLink string
	CreatedAt      time.Time
}

func (b *Booking) Cancel() bool {
	if b.Status == BookingStatusCancelled {
		return false
	}

	b.Status = BookingStatusCancelled

	return true
}

func (b *Booking) IsOwnedBy(userID string) bool {
	return b.UserID == userID
}

type BookingUseCase interface {
	CreateBookingWithConference(
		ctx context.Context,
		userID, slotID string,
	) (*Booking, error)

	CreateBooking(
		ctx context.Context,
		userID, slotID string,
	) (*Booking, error)

	CancelBooking(ctx context.Context, userID, bookingID string) error
	GetMyBookings(ctx context.Context, userID string) ([]*Booking, error)
	GetAllBookings(ctx context.Context, limit, offset int) ([]*Booking, error)
	GetTotalCount(ctx context.Context) (int, error)
}
